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

// ProductRepository define la interfaz para el repositorio de productos
type ProductRepository interface {
	Create(producto *models.Producto) error
	GetByID(productoID string) (*models.Producto, error)
	Update(producto *models.Producto) error
	Delete(productoID string) error
	ListAll() ([]*models.Producto, error)
	GetLowStockProducts() ([]*models.Producto, error)
	UpdateStock(productoID string, nuevaCantidad int) error
}

// productRepository implementa ProductRepository
type productRepository struct {
	db  *database.DynamoDBClient
	log *logrus.Logger
}

// NewProductRepository crea una nueva instancia de ProductRepository
func NewProductRepository(db *database.DynamoDBClient, log *logrus.Logger) ProductRepository {
	return &productRepository{
		db:  db,
		log: log,
	}
}

// Create crea un nuevo producto
func (r *productRepository) Create(producto *models.Producto) error {
	item, err := dynamodbattribute.MarshalMap(producto)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("products"),
		Item:      item,
	}

	_, err = r.db.GetClient().PutItem(input)
	if err != nil {
		r.log.Errorf("Error creating product: %v", err)
		return err
	}

	r.log.Infof("Product created successfully: %s", producto.ProductoID)
	return nil
}

// GetByID obtiene un producto por su ID
func (r *productRepository) GetByID(productoID string) (*models.Producto, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String("products"),
		Key: map[string]*dynamodb.AttributeValue{
			"producto_id": {
				S: aws.String(productoID),
			},
		},
	}

	result, err := r.db.GetClient().GetItem(input)
	if err != nil {
		r.log.Errorf("Error getting product: %v", err)
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var producto models.Producto
	err = dynamodbattribute.UnmarshalMap(result.Item, &producto)
	if err != nil {
		r.log.Errorf("Error unmarshaling product: %v", err)
		return nil, err
	}

	return &producto, nil
}

// Update actualiza un producto existente
func (r *productRepository) Update(producto *models.Producto) error {
	item, err := dynamodbattribute.MarshalMap(producto)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("products"),
		Item:      item,
	}

	_, err = r.db.GetClient().PutItem(input)
	if err != nil {
		r.log.Errorf("Error updating product: %v", err)
		return err
	}

	r.log.Infof("Product updated successfully: %s", producto.ProductoID)
	return nil
}

// Delete elimina un producto
func (r *productRepository) Delete(productoID string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("products"),
		Key: map[string]*dynamodb.AttributeValue{
			"producto_id": {
				S: aws.String(productoID),
			},
		},
	}

	_, err := r.db.GetClient().DeleteItem(input)
	if err != nil {
		r.log.Errorf("Error deleting product: %v", err)
		return err
	}

	r.log.Infof("Product deleted successfully: %s", productoID)
	return nil
}

// ListAll lista todos los productos
func (r *productRepository) ListAll() ([]*models.Producto, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String("products"),
	}

	result, err := r.db.GetClient().Scan(input)
	if err != nil {
		r.log.Errorf("Error scanning products: %v", err)
		return nil, err
	}

	var productos []*models.Producto
	for _, item := range result.Items {
		var producto models.Producto
		err = dynamodbattribute.UnmarshalMap(item, &producto)
		if err != nil {
			r.log.Errorf("Error unmarshaling product: %v", err)
			continue
		}
		productos = append(productos, &producto)
	}

	return productos, nil
}

// GetLowStockProducts obtiene productos con stock bajo
func (r *productRepository) GetLowStockProducts() ([]*models.Producto, error) {
	// Esta implementación requeriría un GSI adicional o un filtro en el scan
	// Por simplicidad, usamos scan con filtro
	filter := expression.LessThanEqual(expression.Name("stock_actual"), expression.Name("punto_reorden"))
	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String("products"),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := r.db.GetClient().Scan(input)
	if err != nil {
		r.log.Errorf("Error scanning low stock products: %v", err)
		return nil, err
	}

	var productos []*models.Producto
	for _, item := range result.Items {
		var producto models.Producto
		err = dynamodbattribute.UnmarshalMap(item, &producto)
		if err != nil {
			r.log.Errorf("Error unmarshaling product: %v", err)
			continue
		}
		productos = append(productos, &producto)
	}

	return productos, nil
}

// UpdateStock actualiza el stock de un producto
func (r *productRepository) UpdateStock(productoID string, nuevaCantidad int) error {
	// Obtener el producto actual
	producto, err := r.GetByID(productoID)
	if err != nil {
		return err
	}

	if producto == nil {
		return nil // Producto no encontrado
	}

	// Actualizar el stock
	producto.StockActual = nuevaCantidad

	// Guardar los cambios
	return r.Update(producto)
}
