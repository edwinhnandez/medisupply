package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"mediplus/purchase-order-service/internal/events"
	"mediplus/purchase-order-service/internal/models"
	"mediplus/purchase-order-service/internal/service"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ExternalEventHandler maneja eventos externos que pueden disparar creación automática de órdenes
type ExternalEventHandler struct {
	orderService service.OrderService
	log          *logrus.Logger
}

// NewExternalEventHandler crea una nueva instancia del handler
func NewExternalEventHandler(orderService service.OrderService, log *logrus.Logger) *ExternalEventHandler {
	return &ExternalEventHandler{
		orderService: orderService,
		log:          log,
	}
}

// HandleExternalEvent procesa eventos externos y determina si debe crear una orden automáticamente
func (h *ExternalEventHandler) HandleExternalEvent(eventData []byte) error {
	h.log.WithField("event_data", string(eventData)).Info("Processing external event")

	// Intentar deserializar como diferentes tipos de eventos externos
	var baseEvent struct {
		EventType string `json:"event_type"`
		Source    string `json:"source"`
	}

	if err := json.Unmarshal(eventData, &baseEvent); err != nil {
		return fmt.Errorf("failed to unmarshal base event: %w", err)
	}

	h.log.WithFields(logrus.Fields{
		"event_type": baseEvent.EventType,
		"source":     baseEvent.Source,
	}).Info("Processing external event")

	switch baseEvent.EventType {
	case events.EventTypeStockBajoExterno:
		return h.handleStockBajoExterno(eventData)
	case events.EventTypeDemandaAltaExterna:
		return h.handleDemandaAltaExterna(eventData)
	case events.EventTypeLoteDanadoExterno:
		return h.handleLoteDanadoExterno(eventData)
	case events.EventTypeAlertaInventarioExterna:
		return h.handleAlertaInventarioExterna(eventData)
	default:
		h.log.WithField("event_type", baseEvent.EventType).Warn("Unknown external event type")
		return nil
	}
}

// handleStockBajoExterno procesa eventos de stock bajo desde sistemas externos
func (h *ExternalEventHandler) handleStockBajoExterno(eventData []byte) error {
	var event events.StockBajoExternoEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to unmarshal StockBajoExternoEvent: %w", err)
	}

	h.log.WithFields(logrus.Fields{
		"producto_id":        event.ProductoID,
		"stock_actual":       event.Data.StockActual,
		"cantidad_requerida": event.Data.CantidadRequerida,
		"prioridad":          event.Data.Prioridad,
		"source":             event.Source,
	}).Info("Processing stock bajo externo event")

	// Crear orden automáticamente
	orden, err := h.createAutoOrderFromStockBajo(&event)
	if err != nil {
		return fmt.Errorf("failed to create auto order from stock bajo: %w", err)
	}

	h.log.WithField("orden_id", orden.OrdenID).Info("Auto order created from stock bajo externo event")
	return nil
}

// handleDemandaAltaExterna procesa eventos de demanda alta desde sistemas externos
func (h *ExternalEventHandler) handleDemandaAltaExterna(eventData []byte) error {
	var event events.DemandaAltaExternaEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to unmarshal DemandaAltaExternaEvent: %w", err)
	}

	h.log.WithFields(logrus.Fields{
		"producto_id":          event.ProductoID,
		"demanda_pronosticada": event.Data.DemandaPronosticada,
		"cantidad_requerida":   event.Data.CantidadRequerida,
		"confianza":            event.Data.ConfianzaPronostico,
		"source":               event.Source,
	}).Info("Processing demanda alta externa event")

	// Solo crear orden si la confianza del pronóstico es alta
	if event.Data.ConfianzaPronostico >= 0.8 {
		orden, err := h.createAutoOrderFromDemandaAlta(&event)
		if err != nil {
			return fmt.Errorf("failed to create auto order from demanda alta: %w", err)
		}

		h.log.WithField("orden_id", orden.OrdenID).Info("Auto order created from demanda alta externa event")
	} else {
		h.log.WithField("confianza", event.Data.ConfianzaPronostico).Info("Skipping auto order creation due to low confidence")
	}

	return nil
}

// handleLoteDanadoExterno procesa eventos de lote dañado desde sistemas externos
func (h *ExternalEventHandler) handleLoteDanadoExterno(eventData []byte) error {
	var event events.LoteDanadoExternoEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to unmarshal LoteDanadoExternoEvent: %w", err)
	}

	h.log.WithFields(logrus.Fields{
		"producto_id":     event.ProductoID,
		"lote_id":         event.Data.LoteID,
		"cantidad_danada": event.Data.CantidadDanada,
		"urgencia":        event.Data.Urgencia,
		"source":          event.Source,
	}).Info("Processing lote danado externo event")

	// Crear orden de reposición automáticamente
	orden, err := h.createAutoOrderFromLoteDanado(&event)
	if err != nil {
		return fmt.Errorf("failed to create auto order from lote danado: %w", err)
	}

	h.log.WithField("orden_id", orden.OrdenID).Info("Auto order created from lote danado externo event")
	return nil
}

// handleAlertaInventarioExterna procesa alertas de inventario desde sistemas externos
func (h *ExternalEventHandler) handleAlertaInventarioExterna(eventData []byte) error {
	var event events.AlertaInventarioExternaEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to unmarshal AlertaInventarioExternaEvent: %w", err)
	}

	h.log.WithFields(logrus.Fields{
		"producto_id":        event.ProductoID,
		"tipo_alerta":        event.Data.TipoAlerta,
		"stock_actual":       event.Data.StockActual,
		"cantidad_requerida": event.Data.CantidadRequerida,
		"prioridad":          event.Data.Prioridad,
		"source":             event.Source,
	}).Info("Processing alerta inventario externa event")

	// Crear orden basada en el tipo de alerta
	orden, err := h.createAutoOrderFromAlertaInventario(&event)
	if err != nil {
		return fmt.Errorf("failed to create auto order from alerta inventario: %w", err)
	}

	h.log.WithField("orden_id", orden.OrdenID).Info("Auto order created from alerta inventario externa event")
	return nil
}

// createAutoOrderFromStockBajo crea una orden automática basada en stock bajo
func (h *ExternalEventHandler) createAutoOrderFromStockBajo(event *events.StockBajoExternoEvent) (*models.OrdenCompra, error) {
	// Buscar proveedor para el producto (por simplicidad, usamos el primer proveedor activo)
	// En un sistema real, esto debería buscar el mejor proveedor para el producto específico
	proveedorID := "dea038d9-d9fa-4cc8-8016-fa20fdc4bcd7" // Proveedor de prueba

	orden := &models.OrdenCompra{
		OrdenID:          uuid.New().String(),
		NumeroOrden:      h.generateOrderNumber("STOCK"),
		ProveedorID:      proveedorID,
		FechaGeneracion:  time.Now(),
		EstadoOrden:      models.EstadoGenerada,
		Prioridad:        h.mapPriority(event.Data.Prioridad),
		MotivoGeneracion: fmt.Sprintf("Stock bajo detectado por %s", event.Source),
		Items: []models.ItemOrdenCompra{
			{
				ItemID:               uuid.New().String(),
				ProductoID:           event.ProductoID,
				CantidadSolicitada:   event.Data.CantidadRequerida,
				PrecioUnitario:       0, // Se calculará después
				TemperaturaRequerida: 0,
				EstadoItem:           models.EstadoItemPendiente,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := h.orderService.CreateOrder(orden)
	if err != nil {
		return nil, err
	}
	return orden, nil
}

// createAutoOrderFromDemandaAlta crea una orden automática basada en demanda alta
func (h *ExternalEventHandler) createAutoOrderFromDemandaAlta(event *events.DemandaAltaExternaEvent) (*models.OrdenCompra, error) {
	proveedorID := "dea038d9-d9fa-4cc8-8016-fa20fdc4bcd7" // Proveedor de prueba

	orden := &models.OrdenCompra{
		OrdenID:          uuid.New().String(),
		NumeroOrden:      h.generateOrderNumber("DEMANDA"),
		ProveedorID:      proveedorID,
		FechaGeneracion:  time.Now(),
		EstadoOrden:      models.EstadoGenerada,
		Prioridad:        h.mapPriority(event.Data.Prioridad),
		MotivoGeneracion: fmt.Sprintf("Demanda alta pronosticada por %s", event.Source),
		Items: []models.ItemOrdenCompra{
			{
				ItemID:               uuid.New().String(),
				ProductoID:           event.ProductoID,
				CantidadSolicitada:   event.Data.CantidadRequerida,
				PrecioUnitario:       0,
				TemperaturaRequerida: 0,
				EstadoItem:           models.EstadoItemPendiente,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := h.orderService.CreateOrder(orden)
	if err != nil {
		return nil, err
	}
	return orden, nil
}

// createAutoOrderFromLoteDanado crea una orden automática basada en lote dañado
func (h *ExternalEventHandler) createAutoOrderFromLoteDanado(event *events.LoteDanadoExternoEvent) (*models.OrdenCompra, error) {
	proveedorID := "dea038d9-d9fa-4cc8-8016-fa20fdc4bcd7" // Proveedor de prueba

	orden := &models.OrdenCompra{
		OrdenID:          uuid.New().String(),
		NumeroOrden:      h.generateOrderNumber("LOTE"),
		ProveedorID:      proveedorID,
		FechaGeneracion:  time.Now(),
		EstadoOrden:      models.EstadoGenerada,
		Prioridad:        models.PrioridadAlta, // Siempre alta prioridad para lotes dañados
		MotivoGeneracion: fmt.Sprintf("Lote dañado detectado por %s", event.Source),
		Items: []models.ItemOrdenCompra{
			{
				ItemID:               uuid.New().String(),
				ProductoID:           event.ProductoID,
				CantidadSolicitada:   event.Data.CantidadDanada,
				PrecioUnitario:       0,
				TemperaturaRequerida: event.Data.TemperaturaRequerida,
				EstadoItem:           models.EstadoItemPendiente,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := h.orderService.CreateOrder(orden)
	if err != nil {
		return nil, err
	}
	return orden, nil
}

// createAutoOrderFromAlertaInventario crea una orden automática basada en alerta de inventario
func (h *ExternalEventHandler) createAutoOrderFromAlertaInventario(event *events.AlertaInventarioExternaEvent) (*models.OrdenCompra, error) {
	proveedorID := "dea038d9-d9fa-4cc8-8016-fa20fdc4bcd7" // Proveedor de prueba

	orden := &models.OrdenCompra{
		OrdenID:          uuid.New().String(),
		NumeroOrden:      h.generateOrderNumber("ALERTA"),
		ProveedorID:      proveedorID,
		FechaGeneracion:  time.Now(),
		EstadoOrden:      models.EstadoGenerada,
		Prioridad:        h.mapPriority(event.Data.Prioridad),
		MotivoGeneracion: fmt.Sprintf("Alerta de inventario desde %s", event.Source),
		Items: []models.ItemOrdenCompra{
			{
				ItemID:               uuid.New().String(),
				ProductoID:           event.ProductoID,
				CantidadSolicitada:   event.Data.CantidadRequerida,
				PrecioUnitario:       0,
				TemperaturaRequerida: 0,
				EstadoItem:           models.EstadoItemPendiente,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := h.orderService.CreateOrder(orden)
	if err != nil {
		return nil, err
	}
	return orden, nil
}

// generateOrderNumber genera un número de orden único
func (h *ExternalEventHandler) generateOrderNumber(prefix string) string {
	return fmt.Sprintf("ORD-%s-%s-%s",
		time.Now().Format("20060102"),
		prefix,
		uuid.New().String()[:8])
}

// mapPriority mapea prioridades de string a enum
func (h *ExternalEventHandler) mapPriority(priority string) models.Prioridad {
	switch priority {
	case "ALTA", "HIGH", "URGENTE":
		return models.PrioridadAlta
	case "MEDIA", "MEDIUM", "NORMAL":
		return models.PrioridadMedia
	case "BAJA", "LOW":
		return models.PrioridadBaja
	default:
		return models.PrioridadMedia
	}
}
