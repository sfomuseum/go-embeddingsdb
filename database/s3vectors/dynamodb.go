package s3vectors

import (
	"context"
	"fmt"

	aa_auth "github.com/aaronland/go-aws/v3/auth"
	aa_dynamodb "github.com/aaronland/go-aws/v3/dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sfomuseum/go-embeddingsdb"
	_ "github.com/sfomuseum/go-embeddingsdb/options"
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

	tables := map[string]*dynamodb.CreateTableInput{
		"s3vectors": &dynamodb.CreateTableInput{
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("Key"), // partition key
					KeyType:       "HASH",
				},
			},
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("Key"),
					AttributeType: "S",
				},
				{
					AttributeName: aws.String("Provider"),
					AttributeType: "S",
				},
			},
			GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
				{
					IndexName: aws.String("by_key"),
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("Key"),
							KeyType:       "HASH",
						},
					},
					Projection: &types.Projection{
						ProjectionType: "ALL",
					},
				},
				{
					IndexName: aws.String("by_provider"),
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("Provider"),
							KeyType:       "HASH",
						},
					},
					Projection: &types.Projection{
						ProjectionType: "ALL",
					},
				},
			},
			BillingMode: types.BillingModePayPerRequest,
			TableName:   aws.String(cl.table),
		},
	}

	table_opts := &aa_dynamodb.CreateTablesOptions{
		Tables: tables,
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

func (cl *DynamoDBClient) ListRecordsByProvider(ctx context.Context, provider string) ([]*DynamoDBRecord, error) {

	cond := expression.Key("Provider").Equal(expression.Value(provider))

	bldr := expression.NewBuilder()
	bldr = bldr.WithKeyCondition(cond)

	expr, err := bldr.Build()

	if err != nil {
		return nil, err
	}

	query_opts := &dynamodb.QueryInput{
		TableName:                 aws.String(cl.table),
		IndexName:                 aws.String("by_provider"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	rsp, err := cl.client.Query(ctx, query_opts)

	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	// Unmarshal the returned items
	var records []*DynamoDBRecord

	err = attributevalue.UnmarshalListOfMaps(rsp.Items, &records)

	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return records, nil
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
