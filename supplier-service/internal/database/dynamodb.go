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
	// Crear tabla de proveedores
	if err := d.createSuppliersTable(); err != nil {
		return err
	}

	// Crear tabla de auditoría
	if err := d.createAuditTable(); err != nil {
		return err
	}

	return nil
}

// createSuppliersTable crea la tabla de proveedores
func (d *DynamoDBClient) createSuppliersTable() error {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String("suppliers"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("proveedor_id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("estado_proveedor"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("proveedor_id"),
				KeyType:       aws.String("HASH"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("estado-index"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("estado_proveedor"),
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

// createAuditTable crea la tabla de auditoría
func (d *DynamoDBClient) createAuditTable() error {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String("audit_traces"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("traza_id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("proveedor_id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("fecha_cambio"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("traza_id"),
				KeyType:       aws.String("HASH"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("proveedor-fecha-index"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("proveedor_id"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("fecha_cambio"),
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
