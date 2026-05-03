package s3vectors

import (
	"context"

	aa_dynamodb "github.com/aaronland/go-aws/v3/dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sfomuseum/go-embeddingsdb"
)

const DynamoDBTableName string = "s3vectors"

type DynamoDBRecord struct {
	Provider    string
	Model       string
	DepictionId string
	Dimensions  int
}

func SetupS3VectorsDynamoDBTable(ctx context.Context, cl *dynamodb.Client) error {

	tables := map[string]*dynamodb.CreateTableInput{
		"s3vectors": &dynamodb.CreateTableInput{
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("Provider"), // partition key
					KeyType:       "HASH",
				},
			},
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("Provider"),
					AttributeType: "S",
				},
				{
					AttributeName: aws.String("Model"),
					AttributeType: "S",
				},
				{
					AttributeName: aws.String("DepictionId"),
					AttributeType: "S",
				},
				{
					AttributeName: aws.String("Dimensions"),
					AttributeType: "N",
				},
			},
			GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
				{
					IndexName: aws.String("by_depiction_id"),
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("DepictionId"),
							KeyType:       "HASH",
						},
					},
					Projection: &types.Projection{
						ProjectionType: "ALL",
					},
				},
				{
					IndexName: aws.String("by_model"),
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("Model"),
							KeyType:       "HASH",
						},
					},
					Projection: &types.Projection{
						ProjectionType: "ALL",
					},
				},
				{
					IndexName: aws.String("by_dimensions"),
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("Dimensions"),
							KeyType:       "HASH",
						},
					},
					Projection: &types.Projection{
						ProjectionType: "KEYS_ONLY",
					},
				},
			},
			BillingMode: types.BillingModePayPerRequest,
			TableName:   aws.String(DynamoDBTableName),
		},
	}

	table_opts := &aa_dynamodb.CreateTablesOptions{
		Tables: tables,
	}

	return aa_dynamodb.CreateTables(ctx, cl, table_opts)
}

func AddRecord(ctx context.Context, cl *dynamodb.Client, rec *embeddingsdb.Record) error {

	dynamodb_rec := recordAsDynamoDBRecord(rec)
	item, err := attributevalue.MarshalMap(dynamodb_rec)

	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(DynamoDBTableName),
		Item:      item,
	}

	_, err = cl.PutItem(ctx, input)
	return err
}

func RemoveRecord(ctx context.Context, cl *dynamodb.Client, rec *embeddingsdb.Record) error {

	return nil
}

func recordAsDynamoDBRecord(rec *embeddingsdb.Record) *DynamoDBRecord {

	return &DynamoDBRecord{
		Model:       rec.Model,
		Provider:    rec.Provider,
		DepictionId: rec.DepictionId,
		Dimensions:  len(rec.Embeddings),
	}
}
