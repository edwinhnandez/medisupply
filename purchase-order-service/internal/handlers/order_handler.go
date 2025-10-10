package handlers

import (
	"mediplus/purchase-order-service/internal/models"
	"mediplus/purchase-order-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// OrderHandler maneja las peticiones HTTP para órdenes
type OrderHandler struct {
	service service.OrderService
	log     *logrus.Logger
}

// NewOrderHandler crea una nueva instancia de OrderHandler
func NewOrderHandler(service service.OrderService, log *logrus.Logger) *OrderHandler {
	return &OrderHandler{
		service: service,
		log:     log,
	}
}

// CreateOrderRequest representa la petición para crear una orden
type CreateOrderRequest struct {
	ProveedorID      string                   `json:"proveedor_id" binding:"required"`
	MotivoGeneracion string                   `json:"motivo_generacion" binding:"required"`
	Prioridad        models.Prioridad         `json:"prioridad" binding:"required"`
	Items            []models.ItemOrdenCompra `json:"items" binding:"required"`
	Evaluacion       *models.Evaluacion       `json:"evaluacion"`
}

// UpdateOrderRequest representa la petición para actualizar una orden
type UpdateOrderRequest struct {
	ProveedorID      string                   `json:"proveedor_id"`
	MotivoGeneracion string                   `json:"motivo_generacion"`
	Prioridad        models.Prioridad         `json:"prioridad"`
	Items            []models.ItemOrdenCompra `json:"items"`
	Evaluacion       *models.Evaluacion       `json:"evaluacion"`
}

// AutoGenerateOrderRequest representa la petición para generar automáticamente una orden
type AutoGenerateOrderRequest struct {
	Trigger string `json:"trigger" binding:"required"`
}

// ProcessStockLowRequest representa la petición para procesar stock bajo
type ProcessStockLowRequest struct {
	ProductoID string `json:"producto_id" binding:"required"`
}

// ProcessLoteDanadoRequest representa la petición para procesar lote dañado
type ProcessLoteDanadoRequest struct {
	ProductoID            string  `json:"producto_id" binding:"required"`
	LoteID                string  `json:"lote_id" binding:"required"`
	CantidadDanada        int     `json:"cantidad_danada" binding:"required"`
	TemperaturaRegistrada float64 `json:"temperatura_registrada" binding:"required"`
}

// ProcessPronosticoDemandaAltaRequest representa la petición para procesar pronóstico de alta demanda
type ProcessPronosticoDemandaAltaRequest struct {
	ProductoID          string `json:"producto_id" binding:"required"`
	DemandaPronosticada int    `json:"demanda_pronosticada" binding:"required"`
}

// CreateOrder crea una nueva orden
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Crear la orden
	orden := models.NewOrdenCompra(req.ProveedorID, req.MotivoGeneracion, req.Prioridad)
	orden.Items = req.Items
	orden.Evaluacion = req.Evaluacion

	err := h.service.CreateOrder(orden)
	if err != nil {
		h.log.Errorf("Error creating order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating order"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"data":    orden,
	})
}

// GetOrder obtiene una orden por ID
func (h *OrderHandler) GetOrder(c *gin.Context) {
	ordenID := c.Param("id")
	if ordenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	orden, err := h.service.GetOrder(ordenID)
	if err != nil {
		h.log.Errorf("Error getting order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting order"})
		return
	}

	if orden == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orden})
}

// UpdateOrder actualiza una orden
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	ordenID := c.Param("id")
	if ordenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	var req UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Obtener la orden actual
	orden, err := h.service.GetOrder(ordenID)
	if err != nil {
		h.log.Errorf("Error getting order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting order"})
		return
	}

	if orden == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Actualizar campos
	if req.ProveedorID != "" {
		orden.ProveedorID = req.ProveedorID
	}
	if req.MotivoGeneracion != "" {
		orden.MotivoGeneracion = req.MotivoGeneracion
	}
	if req.Prioridad != "" {
		orden.Prioridad = req.Prioridad
	}
	if req.Items != nil {
		orden.Items = req.Items
	}
	if req.Evaluacion != nil {
		orden.Evaluacion = req.Evaluacion
	}

	err = h.service.UpdateOrder(orden)
	if err != nil {
		h.log.Errorf("Error updating order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order updated successfully",
		"data":    orden,
	})
}

// DeleteOrder elimina una orden
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	ordenID := c.Param("id")
	if ordenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	err := h.service.DeleteOrder(ordenID)
	if err != nil {
		h.log.Errorf("Error deleting order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}

// ListOrders lista todas las órdenes
func (h *OrderHandler) ListOrders(c *gin.Context) {
	// Obtener parámetros de consulta
	estado := c.Query("estado")
	proveedorID := c.Query("proveedor_id")

	var ordenes []*models.OrdenCompra
	var err error

	if estado != "" {
		// Listar por estado
		estadoOrden := models.EstadoOrden(estado)
		ordenes, err = h.service.ListOrdersByEstado(estadoOrden)
	} else if proveedorID != "" {
		// Listar por proveedor
		ordenes, err = h.service.ListOrdersByProveedor(proveedorID)
	} else {
		// Listar todas
		ordenes, err = h.service.ListOrders()
	}

	if err != nil {
		h.log.Errorf("Error listing orders: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listing orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ordenes})
}

// ConfirmOrder confirma una orden
func (h *OrderHandler) ConfirmOrder(c *gin.Context) {
	ordenID := c.Param("id")
	if ordenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	err := h.service.ConfirmOrder(ordenID)
	if err != nil {
		h.log.Errorf("Error confirming order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error confirming order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order confirmed successfully"})
}

// ReceiveOrder marca una orden como recibida
func (h *OrderHandler) ReceiveOrder(c *gin.Context) {
	ordenID := c.Param("id")
	if ordenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	err := h.service.ReceiveOrder(ordenID)
	if err != nil {
		h.log.Errorf("Error receiving order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error receiving order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order received successfully"})
}

// AutoGenerateOrder genera automáticamente una orden
func (h *OrderHandler) AutoGenerateOrder(c *gin.Context) {
	var req AutoGenerateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.AutoGenerateOrder(req.Trigger)
	if err != nil {
		h.log.Errorf("Error auto-generating order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error auto-generating order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order auto-generated successfully"})
}

// ProcessStockLow procesa un evento de stock bajo
func (h *OrderHandler) ProcessStockLow(c *gin.Context) {
	var req ProcessStockLowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.ProcessStockLowEvent(req.ProductoID)
	if err != nil {
		h.log.Errorf("Error processing stock low event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing stock low event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock low event processed successfully"})
}

// ProcessLoteDanado procesa un evento de lote dañado
func (h *OrderHandler) ProcessLoteDanado(c *gin.Context) {
	var req ProcessLoteDanadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.ProcessLoteDanadoEvent(req.ProductoID, req.LoteID, req.CantidadDanada, req.TemperaturaRegistrada)
	if err != nil {
		h.log.Errorf("Error processing damaged batch event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing damaged batch event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Damaged batch event processed successfully"})
}

// ProcessPronosticoDemandaAlta procesa un evento de pronóstico de alta demanda
func (h *OrderHandler) ProcessPronosticoDemandaAlta(c *gin.Context) {
	var req ProcessPronosticoDemandaAltaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.ProcessPronosticoDemandaAltaEvent(req.ProductoID, req.DemandaPronosticada)
	if err != nil {
		h.log.Errorf("Error processing high demand forecast event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing high demand forecast event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "High demand forecast event processed successfully"})
}
