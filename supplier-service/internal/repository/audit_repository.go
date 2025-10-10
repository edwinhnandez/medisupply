package repository

import (
	"mediplus/supplier-service/internal/database"
	"mediplus/supplier-service/internal/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/sirupsen/logrus"
)

// AuditRepository define la interfaz para el repositorio de auditoría
type AuditRepository interface {
	CreateTraza(traza *models.AuditoriaTraza) error
	GetTrazaByID(trazaID string) (*models.AuditoriaTraza, error)
	GetTrazaByProveedor(proveedorID string) ([]*models.AuditoriaTraza, error)
	GetTrazaByTipoCambio(tipoCambio string) ([]*models.AuditoriaTraza, error)
}

// auditRepository implementa AuditRepository
type auditRepository struct {
	db  *database.DynamoDBClient
	log *logrus.Logger
}

// NewAuditRepository crea una nueva instancia de AuditRepository
func NewAuditRepository(db *database.DynamoDBClient, log *logrus.Logger) AuditRepository {
	return &auditRepository{
		db:  db,
		log: log,
	}
}

// CreateTraza crea una nueva traza de auditoría
func (r *auditRepository) CreateTraza(traza *models.AuditoriaTraza) error {
	item, err := dynamodbattribute.MarshalMap(traza)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("audit_traces"),
		Item:      item,
	}

	_, err = r.db.GetClient().PutItem(input)
	if err != nil {
		r.log.Errorf("Error creating audit trace: %v", err)
		return err
	}

	r.log.Infof("Audit trace created successfully: %s", traza.TrazaID)
	return nil
}

// GetTrazaByID obtiene una traza de auditoría por su ID
func (r *auditRepository) GetTrazaByID(trazaID string) (*models.AuditoriaTraza, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String("audit_traces"),
		Key: map[string]*dynamodb.AttributeValue{
			"traza_id": {
				S: aws.String(trazaID),
			},
		},
	}

	result, err := r.db.GetClient().GetItem(input)
	if err != nil {
		r.log.Errorf("Error getting audit trace: %v", err)
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var traza models.AuditoriaTraza
	err = dynamodbattribute.UnmarshalMap(result.Item, &traza)
	if err != nil {
		r.log.Errorf("Error unmarshaling audit trace: %v", err)
		return nil, err
	}

	return &traza, nil
}

// GetTrazaByProveedor obtiene todas las trazas de auditoría de un proveedor
func (r *auditRepository) GetTrazaByProveedor(proveedorID string) ([]*models.AuditoriaTraza, error) {
	keyCondition := expression.Key("proveedor_id").Equal(expression.Value(proveedorID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String("audit_traces"),
		IndexName:                 aws.String("proveedor-fecha-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(false), // Orden descendente por fecha
	}

	result, err := r.db.GetClient().Query(input)
	if err != nil {
		r.log.Errorf("Error querying audit traces by proveedor: %v", err)
		return nil, err
	}

	var trazas []*models.AuditoriaTraza
	for _, item := range result.Items {
		var traza models.AuditoriaTraza
		err = dynamodbattribute.UnmarshalMap(item, &traza)
		if err != nil {
			r.log.Errorf("Error unmarshaling audit trace: %v", err)
			continue
		}
		trazas = append(trazas, &traza)
	}

	return trazas, nil
}

// GetTrazaByTipoCambio obtiene trazas de auditoría por tipo de cambio
func (r *auditRepository) GetTrazaByTipoCambio(tipoCambio string) ([]*models.AuditoriaTraza, error) {
	filter := expression.Name("tipo_cambio").Equal(expression.Value(tipoCambio))
	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String("audit_traces"),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := r.db.GetClient().Scan(input)
	if err != nil {
		r.log.Errorf("Error scanning audit traces by tipo cambio: %v", err)
		return nil, err
	}

	var trazas []*models.AuditoriaTraza
	for _, item := range result.Items {
		var traza models.AuditoriaTraza
		err = dynamodbattribute.UnmarshalMap(item, &traza)
		if err != nil {
			r.log.Errorf("Error unmarshaling audit trace: %v", err)
			continue
		}
		trazas = append(trazas, &traza)
	}

	return trazas, nil
}
