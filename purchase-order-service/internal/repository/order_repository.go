package repository

import (
	"mediplus/purchase-order-service/internal/database"
	"mediplus/purchase-order-service/internal/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/sirupsen/logrus"
)

// OrderRepository define la interfaz para el repositorio de órdenes
type OrderRepository interface {
	Create(orden *models.OrdenCompra) error
	GetByID(ordenID string) (*models.OrdenCompra, error)
	Update(orden *models.OrdenCompra) error
	Delete(ordenID string) error
	ListByEstado(estado models.EstadoOrden) ([]*models.OrdenCompra, error)
	ListByProveedor(proveedorID string) ([]*models.OrdenCompra, error)
	ListAll() ([]*models.OrdenCompra, error)
	GetByNumeroOrden(numeroOrden string) (*models.OrdenCompra, error)
}

// orderRepository implementa OrderRepository
type orderRepository struct {
	db  *database.DynamoDBClient
	log *logrus.Logger
}

// NewOrderRepository crea una nueva instancia de OrderRepository
func NewOrderRepository(db *database.DynamoDBClient, log *logrus.Logger) OrderRepository {
	return &orderRepository{
		db:  db,
		log: log,
	}
}

// Create crea una nueva orden
func (r *orderRepository) Create(orden *models.OrdenCompra) error {
	item, err := dynamodbattribute.MarshalMap(orden)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("orders"),
		Item:      item,
	}

	_, err = r.db.GetClient().PutItem(input)
	if err != nil {
		r.log.Errorf("Error creating order: %v", err)
		return err
	}

	r.log.Infof("Order created successfully: %s", orden.OrdenID)
	return nil
}

// GetByID obtiene una orden por su ID
func (r *orderRepository) GetByID(ordenID string) (*models.OrdenCompra, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String("orders"),
		Key: map[string]*dynamodb.AttributeValue{
			"orden_id": {
				S: aws.String(ordenID),
			},
		},
	}

	result, err := r.db.GetClient().GetItem(input)
	if err != nil {
		r.log.Errorf("Error getting order: %v", err)
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var orden models.OrdenCompra
	err = dynamodbattribute.UnmarshalMap(result.Item, &orden)
	if err != nil {
		r.log.Errorf("Error unmarshaling order: %v", err)
		return nil, err
	}

	return &orden, nil
}

// Update actualiza una orden existente
func (r *orderRepository) Update(orden *models.OrdenCompra) error {
	item, err := dynamodbattribute.MarshalMap(orden)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("orders"),
		Item:      item,
	}

	_, err = r.db.GetClient().PutItem(input)
	if err != nil {
		r.log.Errorf("Error updating order: %v", err)
		return err
	}

	r.log.Infof("Order updated successfully: %s", orden.OrdenID)
	return nil
}

// Delete elimina una orden
func (r *orderRepository) Delete(ordenID string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("orders"),
		Key: map[string]*dynamodb.AttributeValue{
			"orden_id": {
				S: aws.String(ordenID),
			},
		},
	}

	_, err := r.db.GetClient().DeleteItem(input)
	if err != nil {
		r.log.Errorf("Error deleting order: %v", err)
		return err
	}

	r.log.Infof("Order deleted successfully: %s", ordenID)
	return nil
}

// ListByEstado lista órdenes por estado
func (r *orderRepository) ListByEstado(estado models.EstadoOrden) ([]*models.OrdenCompra, error) {
	keyCondition := expression.Key("estado_orden").Equal(expression.Value(string(estado)))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String("orders"),
		IndexName:                 aws.String("estado-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := r.db.GetClient().Query(input)
	if err != nil {
		r.log.Errorf("Error querying orders by estado: %v", err)
		return nil, err
	}

	var ordenes []*models.OrdenCompra
	for _, item := range result.Items {
		var orden models.OrdenCompra
		err = dynamodbattribute.UnmarshalMap(item, &orden)
		if err != nil {
			r.log.Errorf("Error unmarshaling order: %v", err)
			continue
		}
		ordenes = append(ordenes, &orden)
	}

	return ordenes, nil
}

// ListByProveedor lista órdenes por proveedor
func (r *orderRepository) ListByProveedor(proveedorID string) ([]*models.OrdenCompra, error) {
	keyCondition := expression.Key("proveedor_id").Equal(expression.Value(proveedorID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String("orders"),
		IndexName:                 aws.String("proveedor-fecha-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(false), // Orden descendente por fecha
	}

	result, err := r.db.GetClient().Query(input)
	if err != nil {
		r.log.Errorf("Error querying orders by proveedor: %v", err)
		return nil, err
	}

	var ordenes []*models.OrdenCompra
	for _, item := range result.Items {
		var orden models.OrdenCompra
		err = dynamodbattribute.UnmarshalMap(item, &orden)
		if err != nil {
			r.log.Errorf("Error unmarshaling order: %v", err)
			continue
		}
		ordenes = append(ordenes, &orden)
	}

	return ordenes, nil
}

// ListAll lista todas las órdenes
func (r *orderRepository) ListAll() ([]*models.OrdenCompra, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String("orders"),
	}

	result, err := r.db.GetClient().Scan(input)
	if err != nil {
		r.log.Errorf("Error scanning orders: %v", err)
		return nil, err
	}

	var ordenes []*models.OrdenCompra
	for _, item := range result.Items {
		var orden models.OrdenCompra
		err = dynamodbattribute.UnmarshalMap(item, &orden)
		if err != nil {
			r.log.Errorf("Error unmarshaling order: %v", err)
			continue
		}
		ordenes = append(ordenes, &orden)
	}

	return ordenes, nil
}

// GetByNumeroOrden obtiene una orden por su número de orden
func (r *orderRepository) GetByNumeroOrden(numeroOrden string) (*models.OrdenCompra, error) {
	filter := expression.Name("numero_orden").Equal(expression.Value(numeroOrden))
	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String("orders"),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := r.db.GetClient().Scan(input)
	if err != nil {
		r.log.Errorf("Error scanning orders by numero: %v", err)
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var orden models.OrdenCompra
	err = dynamodbattribute.UnmarshalMap(result.Items[0], &orden)
	if err != nil {
		r.log.Errorf("Error unmarshaling order: %v", err)
		return nil, err
	}

	return &orden, nil
}
