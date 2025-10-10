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

// SupplierRepository define la interfaz para el repositorio de proveedores
type SupplierRepository interface {
	Create(proveedor *models.Proveedor) error
	GetByID(proveedorID string) (*models.Proveedor, error)
	Update(proveedor *models.Proveedor) error
	Delete(proveedorID string) error
	ListByEstado(estado models.EstadoProveedor) ([]*models.Proveedor, error)
	ListAll() ([]*models.Proveedor, error)
	GetByCertificacion(tipoCertificacion string) ([]*models.Proveedor, error)
	GetByCapacidadCadenaFrio() ([]*models.Proveedor, error)
}

// supplierRepository implementa SupplierRepository
type supplierRepository struct {
	db  *database.DynamoDBClient
	log *logrus.Logger
}

// NewSupplierRepository crea una nueva instancia de SupplierRepository
func NewSupplierRepository(db *database.DynamoDBClient, log *logrus.Logger) SupplierRepository {
	return &supplierRepository{
		db:  db,
		log: log,
	}
}

// Create crea un nuevo proveedor
func (r *supplierRepository) Create(proveedor *models.Proveedor) error {
	item, err := dynamodbattribute.MarshalMap(proveedor)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("suppliers"),
		Item:      item,
	}

	_, err = r.db.GetClient().PutItem(input)
	if err != nil {
		r.log.Errorf("Error creating supplier: %v", err)
		return err
	}

	r.log.Infof("Supplier created successfully: %s", proveedor.ProveedorID)
	return nil
}

// GetByID obtiene un proveedor por su ID
func (r *supplierRepository) GetByID(proveedorID string) (*models.Proveedor, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String("suppliers"),
		Key: map[string]*dynamodb.AttributeValue{
			"proveedor_id": {
				S: aws.String(proveedorID),
			},
		},
	}

	result, err := r.db.GetClient().GetItem(input)
	if err != nil {
		r.log.Errorf("Error getting supplier: %v", err)
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var proveedor models.Proveedor
	err = dynamodbattribute.UnmarshalMap(result.Item, &proveedor)
	if err != nil {
		r.log.Errorf("Error unmarshaling supplier: %v", err)
		return nil, err
	}

	return &proveedor, nil
}

// Update actualiza un proveedor existente
func (r *supplierRepository) Update(proveedor *models.Proveedor) error {
	item, err := dynamodbattribute.MarshalMap(proveedor)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("suppliers"),
		Item:      item,
	}

	_, err = r.db.GetClient().PutItem(input)
	if err != nil {
		r.log.Errorf("Error updating supplier: %v", err)
		return err
	}

	r.log.Infof("Supplier updated successfully: %s", proveedor.ProveedorID)
	return nil
}

// Delete elimina un proveedor
func (r *supplierRepository) Delete(proveedorID string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("suppliers"),
		Key: map[string]*dynamodb.AttributeValue{
			"proveedor_id": {
				S: aws.String(proveedorID),
			},
		},
	}

	_, err := r.db.GetClient().DeleteItem(input)
	if err != nil {
		r.log.Errorf("Error deleting supplier: %v", err)
		return err
	}

	r.log.Infof("Supplier deleted successfully: %s", proveedorID)
	return nil
}

// ListByEstado lista proveedores por estado
func (r *supplierRepository) ListByEstado(estado models.EstadoProveedor) ([]*models.Proveedor, error) {
	keyCondition := expression.Key("estado_proveedor").Equal(expression.Value(string(estado)))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String("suppliers"),
		IndexName:                 aws.String("estado-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := r.db.GetClient().Query(input)
	if err != nil {
		r.log.Errorf("Error querying suppliers by estado: %v", err)
		return nil, err
	}

	var proveedores []*models.Proveedor
	for _, item := range result.Items {
		var proveedor models.Proveedor
		err = dynamodbattribute.UnmarshalMap(item, &proveedor)
		if err != nil {
			r.log.Errorf("Error unmarshaling supplier: %v", err)
			continue
		}
		proveedores = append(proveedores, &proveedor)
	}

	return proveedores, nil
}

// ListAll lista todos los proveedores
func (r *supplierRepository) ListAll() ([]*models.Proveedor, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String("suppliers"),
	}

	result, err := r.db.GetClient().Scan(input)
	if err != nil {
		r.log.Errorf("Error scanning suppliers: %v", err)
		return nil, err
	}

	var proveedores []*models.Proveedor
	for _, item := range result.Items {
		var proveedor models.Proveedor
		err = dynamodbattribute.UnmarshalMap(item, &proveedor)
		if err != nil {
			r.log.Errorf("Error unmarshaling supplier: %v", err)
			continue
		}
		proveedores = append(proveedores, &proveedor)
	}

	return proveedores, nil
}

// GetByCertificacion obtiene proveedores por tipo de certificación
func (r *supplierRepository) GetByCertificacion(tipoCertificacion string) ([]*models.Proveedor, error) {
	// Esta implementación requeriría un GSI adicional o un filtro en el scan
	// Por simplicidad, usamos scan con filtro
	filter := expression.Contains(expression.Name("certificaciones"), tipoCertificacion)
	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String("suppliers"),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := r.db.GetClient().Scan(input)
	if err != nil {
		r.log.Errorf("Error scanning suppliers by certification: %v", err)
		return nil, err
	}

	var proveedores []*models.Proveedor
	for _, item := range result.Items {
		var proveedor models.Proveedor
		err = dynamodbattribute.UnmarshalMap(item, &proveedor)
		if err != nil {
			r.log.Errorf("Error unmarshaling supplier: %v", err)
			continue
		}
		proveedores = append(proveedores, &proveedor)
	}

	return proveedores, nil
}

// GetByCapacidadCadenaFrio obtiene proveedores con capacidad de cadena de frío
func (r *supplierRepository) GetByCapacidadCadenaFrio() ([]*models.Proveedor, error) {
	filter := expression.AttributeExists(expression.Name("capacidad_logistica.capacidad_cadena_frio")).And(
		expression.Name("capacidad_logistica.capacidad_cadena_frio").Equal(expression.Value(true)))

	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String("suppliers"),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := r.db.GetClient().Scan(input)
	if err != nil {
		r.log.Errorf("Error scanning suppliers by cold chain capacity: %v", err)
		return nil, err
	}

	var proveedores []*models.Proveedor
	for _, item := range result.Items {
		var proveedor models.Proveedor
		err = dynamodbattribute.UnmarshalMap(item, &proveedor)
		if err != nil {
			r.log.Errorf("Error unmarshaling supplier: %v", err)
			continue
		}
		proveedores = append(proveedores, &proveedor)
	}

	return proveedores, nil
}
