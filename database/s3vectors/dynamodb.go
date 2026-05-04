// Package s3vectors contains helper code for maintaining a DynamoDB table
// that mirrors the records stored in an S3 Vectors index.  The table
// is used by the main database implementation to enable efficient
// listing by provider or model.
package s3vectors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	aa_auth "github.com/aaronland/go-aws/v3/auth"
	aa_dynamodb "github.com/aaronland/go-aws/v3/dynamodb"
	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/cursor"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/options"
)

// DynamoDBTableName is the default name of the table used to store
// the record key, provider, model, depiction id and dimension count.
const DynamoDBTableName string = "s3vectors"

// DynamoDBRecord represents a minimal record stored in DynamoDB.
// The primary key is the concatenated provider|model|depiction_id.
type DynamoDBRecord struct {
	Key         string
	Provider    string
	Model       string
	DepictionId string
	Dimensions  int
}

// DynamoDBClient wraps basic add and remove operations the DynamoDB table
// used to mirror the S3 Vectors index.
type DynamoDBClient struct {
	client *dynamodb.Client
	table  string
}

// NewDynamoDBClient creates a DynamoDB client using the AWS
// configuration supplied by cfg_uri.
func NewDynamoDBClient(ctx context.Context, cfg_uri string) (*DynamoDBClient, error) {

	dynamodb_table := DynamoDBTableName

	u, err := url.Parse(cfg_uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	if q.Has("dynamodb-table") {

		v := q.Get("dynamodb-table")

		if v == "" {
			return nil, fmt.Errorf("?dynamodb-table= parameter may not be empty.")
		}

		dynamodb_table = v
	}

	cfg, err := aa_auth.NewConfig(ctx, cfg_uri)

	if err != nil {
		return nil, err
	}

	dynamodb_cl := dynamodb.NewFromConfig(cfg)

	cl := &DynamoDBClient{
		client: dynamodb_cl,
		table:  dynamodb_table,
	}

	return cl, nil
}

// SetupTable creates the DynamoDB table if it does not already exist.
func (cl *DynamoDBClient) SetupTable(ctx context.Context) error {

	table_opts := &aa_dynamodb.CreateTablesOptions{
		Tables: DynamoDBTables(cl.table),
	}

	return aa_dynamodb.CreateTables(ctx, cl.client, table_opts)
}

// AddRecord inserts the supplied Record into the DynamoDB table.
// The function serialises the Record into a DynamoDBRecord
// and marshals it into a DynamoDB item.
func (cl *DynamoDBClient) AddRecord(ctx context.Context, rec *embeddingsdb.Record) error {

	dynamodb_rec := recordAsDynamoDBRecord(rec)
	item, err := attributevalue.MarshalMap(dynamodb_rec)

	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(cl.table),
		Item:      item,
	}

	_, err = cl.client.PutItem(ctx, input)
	return err
}

// RemoveRecord deletes the record identified by rec from DynamoDB.
func (cl *DynamoDBClient) RemoveRecord(ctx context.Context, rec *embeddingsdb.Record) error {

	delete_opts := &dynamodb.DeleteItemInput{
		TableName: aws.String(DynamoDBTableName),
		Key: map[string]types.AttributeValue{
			"Key": &types.AttributeValueMemberS{Value: rec.Key()},
		},
	}

	_, err := cl.client.DeleteItem(ctx, delete_opts)

	return err
}

// ListRecordsByProvider returns all records that match the given provider.
// The results are paginated according to pg_opts.  An optional
// model filter may be supplied via options.
func (cl *DynamoDBClient) ListRecordsByProvider(ctx context.Context, pg_opts pagination.Options, provider string, custom_opts ...options.Option) ([]*DynamoDBRecord, pagination.Results, error) {

	cond := expression.Key("Provider").Equal(expression.Value(provider))

	bldr := expression.NewBuilder()
	bldr = bldr.WithKeyCondition(cond)

	model := options.GetModelFromOptions(ctx, custom_opts...)

	if model != nil {
		filt := expression.Name("Model").Equal(expression.Value(*model))
		bldr = bldr.WithFilter(filt)
	}

	expr, err := bldr.Build()

	if err != nil {
		return nil, nil, err
	}

	query_opts := &dynamodb.QueryInput{
		TableName:                 aws.String(cl.table),
		IndexName:                 aws.String("by_provider_model"),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	return cl.queryRecords(ctx, query_opts, pg_opts)
}

// ListRecordsByModel returns all records that match the given model.
// The results are paginated according to pg_opts.  An optional
// provider filter may be supplied via options.
func (cl *DynamoDBClient) ListRecordsByModel(ctx context.Context, pg_opts pagination.Options, model string, custom_opts ...options.Option) ([]*DynamoDBRecord, pagination.Results, error) {

	cond := expression.Key("Model").Equal(expression.Value(model))

	bldr := expression.NewBuilder()
	bldr = bldr.WithKeyCondition(cond)

	provider := options.GetModelFromOptions(ctx, custom_opts...)

	if provider != nil {
		filt := expression.Name("Provider").Equal(expression.Value(*provider))
		bldr = bldr.WithFilter(filt)
	}

	expr, err := bldr.Build()

	if err != nil {
		return nil, nil, err
	}

	query_opts := &dynamodb.QueryInput{
		TableName:                 aws.String(cl.table),
		IndexName:                 aws.String("by_model_provider"),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	return cl.queryRecords(ctx, query_opts, pg_opts)
}

// queryRecords executes a DynamoDB query and returns the decoded
// records together with pagination cursors.
func (cl *DynamoDBClient) queryRecords(ctx context.Context, query_opts *dynamodb.QueryInput, pg_opts pagination.Options) ([]*DynamoDBRecord, pagination.Results, error) {

	per_page := pg_opts.PerPage()
	pointer := pg_opts.Pointer()

	var prev_cursor string
	var next_cursor string

	if per_page > 0 {
		query_opts.Limit = aws.Int32(int32(per_page))
	}

	if pointer != nil {

		str_key, ok := pointer.(string)

		if ok && str_key != "" {

			str_key = strings.Replace(str_key, "after-", "", 1)
			str_key = strings.Replace(str_key, "before-", "", 1)

			start_key, err := decodeStartKey(str_key)

			if err != nil {
				slog.Warn("Failed to unmarshal start key", "error", err, "key", str_key)
			} else {
				query_opts.ExclusiveStartKey = start_key
			}
		}

	}

	rsp, err := cl.client.Query(ctx, query_opts)

	if err != nil {
		return nil, nil, fmt.Errorf("query: %w", err)
	}

	if rsp.LastEvaluatedKey != nil {

		enc_key, err := encodeStartKey(rsp.LastEvaluatedKey)

		if err != nil {
			return nil, nil, err
		}

		next_cursor = enc_key
	}

	// Unmarshal the returned items
	var records []*DynamoDBRecord

	err = attributevalue.UnmarshalListOfMaps(rsp.Items, &records)

	if err != nil {
		return nil, nil, fmt.Errorf("unmarshal: %w", err)
	}

	pg_rsp, err := cursor.NewPaginationFromCursors(prev_cursor, next_cursor)

	if err != nil {
		return nil, nil, err
	}

	return records, pg_rsp, nil
}

// recordAsDynamoDBRecord converts an embeddingsdb.Record into a DynamoDBRecord.
func recordAsDynamoDBRecord(rec *embeddingsdb.Record) *DynamoDBRecord {

	return &DynamoDBRecord{
		Key:         rec.Key(),
		Model:       rec.Model,
		Provider:    rec.Provider,
		DepictionId: rec.DepictionId,
		Dimensions:  len(rec.Embeddings),
	}
}

// encodeStartKey serialises a DynamoDB key map into a base64 string
// suitable for use as a pagination cursor.
func encodeStartKey(key map[string]types.AttributeValue) (string, error) {

	if len(key) == 0 {
		return "", fmt.Errorf("Missing key")
	}

	var plain_map map[string]interface{}

	err := attributevalue.UnmarshalMap(key, &plain_map)

	if err != nil {
		return "", err
	}

	data, err := json.Marshal(plain_map)

	if err != nil {
		return "", err
	}

	enc := base64.URLEncoding.EncodeToString(data)
	return enc, nil
}

// decodeStartKey deserialises a pagination cursor back into
// a DynamoDB key map.
func decodeStartKey(str_key string) (map[string]types.AttributeValue, error) {

	if str_key == "" {
		return nil, fmt.Errorf("Missing key")
	}

	data, err := base64.URLEncoding.DecodeString(str_key)

	if err != nil {

		slog.Error("WTF", "key", str_key, "error", err)
		return nil, fmt.Errorf("Failed to decode key, %w", err)
	}

	var plain_map map[string]interface{}

	err = json.Unmarshal(data, &plain_map)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal map, %w", err)
	}

	start_key, err := attributevalue.MarshalMap(plain_map)

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal start key, %w", err)
	}

	return start_key, nil
}
