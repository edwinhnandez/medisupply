package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/sirupsen/logrus"
)

// DynamoDBClient encapsula el cliente de DynamoDB
type DynamoDBClient struct {
	client *dynamodb.DynamoDB
	log    *logrus.Logger
}

// NewDynamoDBClient crea una nueva instancia del cliente DynamoDB
func NewDynamoDBClient(region, endpoint string) (*DynamoDBClient, error) {
	config := &aws.Config{
		Region: aws.String(region),
	}

	// Si se proporciona un endpoint (para desarrollo local), usarlo
	if endpoint != "" {
		config.Endpoint = aws.String(endpoint)
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := dynamodb.New(sess)

	return &DynamoDBClient{
		client: client,
		log:    logrus.New(),
	}, nil
}

// GetClient retorna el cliente de DynamoDB
func (d *DynamoDBClient) GetClient() *dynamodb.DynamoDB {
	return d.client
}

// CreateTables crea las tablas necesarias en DynamoDB
func (d *DynamoDBClient) CreateTables() error {
	// Crear tabla de órdenes
	if err := d.createOrdersTable(); err != nil {
		return err
	}

	// Crear tabla de productos
	if err := d.createProductsTable(); err != nil {
		return err
	}

	return nil
}

// createOrdersTable crea la tabla de órdenes
func (d *DynamoDBClient) createOrdersTable() error {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String("orders"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("orden_id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("estado_orden"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("proveedor_id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("fecha_generacion"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("orden_id"),
				KeyType:       aws.String("HASH"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("estado-index"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("estado_orden"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
			{
				IndexName: aws.String("proveedor-fecha-index"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("proveedor_id"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("fecha_generacion"),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := d.client.CreateTable(input)
	if err != nil {
		// Si la tabla ya existe, no es un error
		if _, ok := err.(*dynamodb.ResourceInUseException); !ok {
			return err
		}
	}

	return nil
}

// createProductsTable crea la tabla de productos
func (d *DynamoDBClient) createProductsTable() error {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String("products"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("producto_id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("stock_actual"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("producto_id"),
				KeyType:       aws.String("HASH"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("stock-index"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("stock_actual"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := d.client.CreateTable(input)
	if err != nil {
		// Si la tabla ya existe, no es un error
		if _, ok := err.(*dynamodb.ResourceInUseException); !ok {
			return err
		}
	}

	return nil
}
