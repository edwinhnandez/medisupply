package service

import (
	"mediplus/purchase-order-service/internal/events"
	"mediplus/purchase-order-service/internal/models"
	"mediplus/purchase-order-service/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// OrderService define la interfaz para el servicio de órdenes
type OrderService interface {
	CreateOrder(orden *models.OrdenCompra) error
	GetOrder(ordenID string) (*models.OrdenCompra, error)
	UpdateOrder(orden *models.OrdenCompra) error
	DeleteOrder(ordenID string) error
	ListOrders() ([]*models.OrdenCompra, error)
	ConfirmOrder(ordenID string) error
	ReceiveOrder(ordenID string) error
	AutoGenerateOrder(trigger string) error
	GetOrderByNumero(numeroOrden string) (*models.OrdenCompra, error)
	ProcessStockLowEvent(productoID string) error
	ProcessLoteDanadoEvent(productoID, loteID string, cantidadDanada int, temperaturaRegistrada float64) error
	ProcessPronosticoDemandaAltaEvent(productoID string, demandaPronosticada int) error
	ListOrdersByEstado(estado models.EstadoOrden) ([]*models.OrdenCompra, error)
	ListOrdersByProveedor(proveedorID string) ([]*models.OrdenCompra, error)
}

// orderService implementa OrderService
type orderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	eventBus    events.EventBus
	log         *logrus.Logger
}

// NewOrderService crea una nueva instancia de OrderService
func NewOrderService(
	orderRepo repository.OrderRepository,
	productRepo repository.ProductRepository,
	eventBus events.EventBus,
	log *logrus.Logger,
) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		eventBus:    eventBus,
		log:         log,
	}
}

// CreateOrder crea una nueva orden
func (s *orderService) CreateOrder(orden *models.OrdenCompra) error {
	// Crear la orden
	err := s.orderRepo.Create(orden)
	if err != nil {
		s.log.Errorf("Error creating order: %v", err)
		return err
	}

	// Emitir evento de orden generada
	event := &events.OrdenCompraGeneradaEvent{
		EventID:   uuid.New().String(),
		EventType: events.EventTypeOrdenCompraGenerada,
		OrdenID:   orden.OrdenID,
		Timestamp: time.Now(),
	}

	event.Data.NumeroOrden = orden.NumeroOrden
	event.Data.ProveedorID = orden.ProveedorID
	event.Data.MotivoGeneracion = orden.MotivoGeneracion
	event.Data.Prioridad = string(orden.Prioridad)
	event.Data.TotalItems = len(orden.Items)

	// Calcular valor total
	var valorTotal float64
	for _, item := range orden.Items {
		valorTotal += item.PrecioUnitario * float64(item.CantidadSolicitada)
	}
	event.Data.ValorTotal = valorTotal

	err = s.eventBus.Publish(events.TopicOrderEvents, event)
	if err != nil {
		s.log.Errorf("Error publishing order generated event: %v", err)
	}

	return nil
}

// GetOrder obtiene una orden por su ID
func (s *orderService) GetOrder(ordenID string) (*models.OrdenCompra, error) {
	return s.orderRepo.GetByID(ordenID)
}

// UpdateOrder actualiza una orden
func (s *orderService) UpdateOrder(orden *models.OrdenCompra) error {
	return s.orderRepo.Update(orden)
}

// DeleteOrder elimina una orden
func (s *orderService) DeleteOrder(ordenID string) error {
	return s.orderRepo.Delete(ordenID)
}

// ListOrders lista todas las órdenes
func (s *orderService) ListOrders() ([]*models.OrdenCompra, error) {
	return s.orderRepo.ListAll()
}

// ConfirmOrder confirma una orden
func (s *orderService) ConfirmOrder(ordenID string) error {
	// Obtener la orden
	orden, err := s.orderRepo.GetByID(ordenID)
	if err != nil {
		return err
	}

	if orden == nil {
		return nil // Orden no encontrada
	}

	// Confirmar la orden
	orden.ConfirmOrder()

	// Actualizar en la base de datos
	err = s.orderRepo.Update(orden)
	if err != nil {
		s.log.Errorf("Error updating confirmed order: %v", err)
		return err
	}

	// Emitir evento de orden confirmada
	event := &events.OrdenCompraConfirmadaEvent{
		EventID:   uuid.New().String(),
		EventType: events.EventTypeOrdenCompraConfirmada,
		OrdenID:   ordenID,
		Timestamp: time.Now(),
	}

	event.Data.NumeroOrden = orden.NumeroOrden
	event.Data.ProveedorID = orden.ProveedorID
	event.Data.FechaConfirmacion = time.Now()

	err = s.eventBus.Publish(events.TopicOrderEvents, event)
	if err != nil {
		s.log.Errorf("Error publishing order confirmed event: %v", err)
	}

	return nil
}

// ReceiveOrder marca una orden como recibida
func (s *orderService) ReceiveOrder(ordenID string) error {
	// Obtener la orden
	orden, err := s.orderRepo.GetByID(ordenID)
	if err != nil {
		return err
	}

	if orden == nil {
		return nil // Orden no encontrada
	}

	// Marcar como recibida
	orden.ReceiveOrder()

	// Actualizar en la base de datos
	err = s.orderRepo.Update(orden)
	if err != nil {
		s.log.Errorf("Error updating received order: %v", err)
		return err
	}

	// Emitir evento de orden recibida
	event := &events.OrdenCompraRecibidaEvent{
		EventID:   uuid.New().String(),
		EventType: events.EventTypeOrdenCompraRecibida,
		OrdenID:   ordenID,
		Timestamp: time.Now(),
	}

	event.Data.NumeroOrden = orden.NumeroOrden
	event.Data.ProveedorID = orden.ProveedorID
	event.Data.FechaRecepcion = time.Now()

	err = s.eventBus.Publish(events.TopicOrderEvents, event)
	if err != nil {
		s.log.Errorf("Error publishing order received event: %v", err)
	}

	return nil
}

// AutoGenerateOrder genera automáticamente una orden basada en el trigger
func (s *orderService) AutoGenerateOrder(trigger string) error {
	s.log.Infof("Auto-generating order for trigger: %s", trigger)

	// Obtener productos con stock bajo
	productos, err := s.productRepo.GetLowStockProducts()
	if err != nil {
		s.log.Errorf("Error getting low stock products: %v", err)
		return err
	}

	if len(productos) == 0 {
		s.log.Info("No products with low stock found")
		return nil
	}

	// Por simplicidad, crear una orden para el primer producto
	// En un escenario real, se implementaría lógica más compleja
	producto := productos[0]

	// Calcular cantidad a solicitar
	cantidadRequerida := producto.StockMaximo - producto.StockActual

	// Crear orden
	orden := models.NewOrdenCompra("", trigger, models.PrioridadMedia)
	orden.MotivoGeneracion = trigger

	// Agregar item a la orden
	item := models.NewItemOrdenCompra(
		producto.ProductoID,
		cantidadRequerida,
		100.0, // Precio unitario por defecto
		producto.Condiciones.TemperaturaMinima,
	)
	orden.AddItem(item)

	// Crear la orden
	err = s.CreateOrder(orden)
	if err != nil {
		s.log.Errorf("Error creating auto-generated order: %v", err)
		return err
	}

	s.log.Infof("Auto-generated order created: %s", orden.OrdenID)
	return nil
}

// GetOrderByNumero obtiene una orden por su número
func (s *orderService) GetOrderByNumero(numeroOrden string) (*models.OrdenCompra, error) {
	return s.orderRepo.GetByNumeroOrden(numeroOrden)
}

// ProcessStockLowEvent procesa un evento de stock bajo
func (s *orderService) ProcessStockLowEvent(productoID string) error {
	s.log.Infof("Processing stock low event for product: %s", productoID)

	// Obtener el producto
	producto, err := s.productRepo.GetByID(productoID)
	if err != nil {
		s.log.Errorf("Error getting product: %v", err)
		return err
	}

	if producto == nil {
		s.log.Warnf("Product not found: %s", productoID)
		return nil
	}

	// Verificar si ya existe una orden pendiente para este producto
	ordenesPendientes, err := s.orderRepo.ListByEstado(models.EstadoGenerada)
	if err != nil {
		s.log.Errorf("Error getting pending orders: %v", err)
		return err
	}

	// Verificar si ya hay una orden pendiente para este producto
	for _, orden := range ordenesPendientes {
		for _, item := range orden.Items {
			if item.ProductoID == productoID {
				s.log.Infof("Order already exists for product %s: %s", productoID, orden.OrdenID)
				return nil
			}
		}
	}

	// Crear orden automáticamente para el producto con stock bajo
	return s.createOrderForLowStockProduct(producto)
}

// ProcessLoteDanadoEvent procesa un evento de lote dañado
func (s *orderService) ProcessLoteDanadoEvent(productoID, loteID string, cantidadDanada int, temperaturaRegistrada float64) error {
	s.log.Infof("Processing damaged batch event for product: %s, lot: %s", productoID, loteID)

	// Obtener el producto
	producto, err := s.productRepo.GetByID(productoID)
	if err != nil {
		s.log.Errorf("Error getting product: %v", err)
		return err
	}

	if producto == nil {
		s.log.Warnf("Product not found: %s", productoID)
		return nil
	}

	// Emitir evento de lote dañado
	event := &events.LoteDanadoEvent{
		EventID:    uuid.New().String(),
		EventType:  events.EventTypeLoteDanado,
		ProductoID: productoID,
		Timestamp:  time.Now(),
	}

	event.Data.NombreProducto = producto.Nombre
	event.Data.LoteID = loteID
	event.Data.CantidadDanada = cantidadDanada
	event.Data.TemperaturaRegistrada = temperaturaRegistrada
	if producto.Condiciones != nil {
		event.Data.TemperaturaRequerida = producto.Condiciones.TemperaturaMinima
	}
	event.Data.MotivoDanio = "Temperatura fuera de rango"

	err = s.eventBus.Publish(events.TopicStockEvents, event)
	if err != nil {
		s.log.Errorf("Error publishing damaged batch event: %v", err)
	}

	// Auto-generar orden para reponer el stock dañado
	return s.AutoGenerateOrder("Lote dañado por temperatura")
}

// ProcessPronosticoDemandaAltaEvent procesa un evento de pronóstico de alta demanda
func (s *orderService) ProcessPronosticoDemandaAltaEvent(productoID string, demandaPronosticada int) error {
	s.log.Infof("Processing high demand forecast event for product: %s", productoID)

	// Obtener el producto
	producto, err := s.productRepo.GetByID(productoID)
	if err != nil {
		s.log.Errorf("Error getting product: %v", err)
		return err
	}

	if producto == nil {
		s.log.Warnf("Product not found: %s", productoID)
		return nil
	}

	// Emitir evento de pronóstico de alta demanda
	event := &events.PronosticoDemandaAltaEvent{
		EventID:    uuid.New().String(),
		EventType:  events.EventTypePronosticoDemandaAlta,
		ProductoID: productoID,
		Timestamp:  time.Now(),
	}

	event.Data.NombreProducto = producto.Nombre
	event.Data.DemandaPronosticada = demandaPronosticada
	event.Data.StockActual = producto.StockActual
	event.Data.CantidadRequerida = demandaPronosticada - producto.StockActual
	event.Data.ConfianzaPronostico = 0.85 // Valor por defecto

	err = s.eventBus.Publish(events.TopicStockEvents, event)
	if err != nil {
		s.log.Errorf("Error publishing high demand forecast event: %v", err)
	}

	// Auto-generar orden si la demanda pronosticada excede el stock actual
	if demandaPronosticada > producto.StockActual {
		return s.AutoGenerateOrder("Pronóstico demanda alta")
	}

	return nil
}

// ListOrdersByEstado lista órdenes por estado
func (s *orderService) ListOrdersByEstado(estado models.EstadoOrden) ([]*models.OrdenCompra, error) {
	return s.orderRepo.ListByEstado(estado)
}

// ListOrdersByProveedor lista órdenes por proveedor
func (s *orderService) ListOrdersByProveedor(proveedorID string) ([]*models.OrdenCompra, error) {
	return s.orderRepo.ListByProveedor(proveedorID)
}

// createOrderForLowStockProduct crea una orden automáticamente para un producto con stock bajo
func (s *orderService) createOrderForLowStockProduct(producto *models.Producto) error {
	s.log.Infof("Creating automatic order for low stock product: %s", producto.ProductoID)

	// Calcular cantidad a solicitar (hasta el stock máximo)
	cantidadRequerida := producto.StockMaximo - producto.StockActual

	if cantidadRequerida <= 0 {
		s.log.Infof("No reorder needed for product %s (stock already at maximum)", producto.ProductoID)
		return nil
	}

	// Determinar prioridad basada en el nivel de stock
	var prioridad models.Prioridad
	if producto.StockActual == 0 {
		prioridad = models.PrioridadCritica
	} else if producto.StockActual <= producto.PuntoReorden/2 {
		prioridad = models.PrioridadAlta
	} else {
		prioridad = models.PrioridadMedia
	}

	// Crear orden
	orden := models.NewOrdenCompra("", "Stock bajo punto reorden", prioridad)
	orden.MotivoGeneracion = "Stock bajo punto reorden - Generación automática"

	// Agregar item a la orden
	precioUnitario := 100.0 // Precio por defecto - en un escenario real se obtendría del catálogo
	temperaturaRequerida := 0.0
	if producto.Condiciones != nil {
		temperaturaRequerida = producto.Condiciones.TemperaturaMinima
	}

	item := models.NewItemOrdenCompra(
		producto.ProductoID,
		cantidadRequerida,
		precioUnitario,
		temperaturaRequerida,
	)
	orden.AddItem(item)

	// Crear la orden
	err := s.CreateOrder(orden)
	if err != nil {
		s.log.Errorf("Error creating automatic order for low stock product: %v", err)
		return err
	}

	s.log.WithFields(logrus.Fields{
		"orden_id":      orden.OrdenID,
		"producto_id":   producto.ProductoID,
		"cantidad":      cantidadRequerida,
		"prioridad":     prioridad,
		"stock_actual":  producto.StockActual,
		"punto_reorden": producto.PuntoReorden,
	}).Info("Successfully created automatic order for low stock product")

	return nil
}
