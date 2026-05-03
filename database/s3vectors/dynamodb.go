package s3vectors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"

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

const DynamoDBTableName string = "s3vectors"

type DynamoDBRecord struct {
	Key         string
	Provider    string
	Model       string
	DepictionId string
	Dimensions  int
}

type DynamoDBClient struct {
	client *dynamodb.Client
	table  string
}

func NewDynamoDBClient(ctx context.Context, cfg_uri string) (*DynamoDBClient, error) {

	cfg, err := aa_auth.NewConfig(ctx, cfg_uri)

	if err != nil {
		return nil, err
	}

	dynamodb_cl := dynamodb.NewFromConfig(cfg)

	cl := &DynamoDBClient{
		client: dynamodb_cl,
		table:  DynamoDBTableName,
	}

	return cl, nil
}

func (cl *DynamoDBClient) SetupTable(ctx context.Context) error {

	table_opts := &aa_dynamodb.CreateTablesOptions{
		Tables: DynamoDBTables(cl.table),
	}

	return aa_dynamodb.CreateTables(ctx, cl.client, table_opts)
}

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

			start_key, err := decodeStartKey(str_key)

			if err != nil {
				slog.Warn("Failed to unmarshal start key", "error", err)
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

func recordAsDynamoDBRecord(rec *embeddingsdb.Record) *DynamoDBRecord {

	return &DynamoDBRecord{
		Key:         rec.Key(),
		Model:       rec.Model,
		Provider:    rec.Provider,
		DepictionId: rec.DepictionId,
		Dimensions:  len(rec.Embeddings),
	}
}

func encodeStartKey(key map[string]types.AttributeValue) (string, error) {

	if len(key) == 0 {
		return "", nil
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

	return base64.URLEncoding.EncodeToString(data), nil
}

func decodeStartKey(str_key string) (map[string]types.AttributeValue, error) {

	if str_key == "" {
		return nil, nil
	}

	data, err := base64.URLEncoding.DecodeString(str_key)

	if err != nil {
		return nil, err
	}

	var plain_map map[string]interface{}

	err = json.Unmarshal(data, &plain_map)

	if err != nil {
		return nil, err
	}

	start_key, err := attributevalue.MarshalMap(plain_map)

	if err != nil {
		return nil, err
	}

	return start_key, nil
}
