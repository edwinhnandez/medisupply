package handlers

import (
	"net/http"
	"time"

	"mediplus/purchase-order-service/internal/events"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ExternalSimulatorHandler maneja la simulación de eventos externos
type ExternalSimulatorHandler struct {
	eventBus events.EventBus
	log      *logrus.Logger
}

// NewExternalSimulatorHandler crea una nueva instancia del handler
func NewExternalSimulatorHandler(eventBus events.EventBus, log *logrus.Logger) *ExternalSimulatorHandler {
	return &ExternalSimulatorHandler{
		eventBus: eventBus,
		log:      log,
	}
}

// SimulateStockBajoExterno simula un evento de stock bajo desde un sistema externo
func (h *ExternalSimulatorHandler) SimulateStockBajoExterno(c *gin.Context) {
	var request struct {
		ProductoID        string `json:"producto_id" binding:"required"`
		NombreProducto    string `json:"nombre_producto" binding:"required"`
		StockActual       int    `json:"stock_actual" binding:"required"`
		PuntoReorden      int    `json:"punto_reorden" binding:"required"`
		StockMaximo       int    `json:"stock_maximo" binding:"required"`
		CantidadRequerida int    `json:"cantidad_requerida" binding:"required"`
		Prioridad         string `json:"prioridad"`
		Urgencia          string `json:"urgencia"`
		Source            string `json:"source"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Valores por defecto
	if request.Prioridad == "" {
		request.Prioridad = "ALTA"
	}
	if request.Urgencia == "" {
		request.Urgencia = "MEDIA"
	}
	if request.Source == "" {
		request.Source = "Sistema de Inventario Externo"
	}

	event := events.StockBajoExternoEvent{
		EventID:    uuid.New().String(),
		EventType:  events.EventTypeStockBajoExterno,
		ProductoID: request.ProductoID,
		Timestamp:  time.Now(),
		Source:     request.Source,
		Data: struct {
			NombreProducto    string `json:"nombre_producto"`
			StockActual       int    `json:"stock_actual"`
			PuntoReorden      int    `json:"punto_reorden"`
			StockMaximo       int    `json:"stock_maximo"`
			CantidadRequerida int    `json:"cantidad_requerida"`
			Prioridad         string `json:"prioridad"`
			Urgencia          string `json:"urgencia"`
		}{
			NombreProducto:    request.NombreProducto,
			StockActual:       request.StockActual,
			PuntoReorden:      request.PuntoReorden,
			StockMaximo:       request.StockMaximo,
			CantidadRequerida: request.CantidadRequerida,
			Prioridad:         request.Prioridad,
			Urgencia:          request.Urgencia,
		},
	}

	// Publicar evento
	err := h.eventBus.Publish(events.TopicExternalEvents, &event)
	if err != nil {
		h.log.WithError(err).Error("Failed to publish stock bajo externo event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
		return
	}

	h.log.WithFields(logrus.Fields{
		"producto_id":        request.ProductoID,
		"stock_actual":       request.StockActual,
		"cantidad_requerida": request.CantidadRequerida,
		"source":             request.Source,
	}).Info("Stock bajo externo event published")

	c.JSON(http.StatusOK, gin.H{
		"message":    "Stock bajo externo event published successfully",
		"event_id":   event.EventID,
		"event_type": event.EventType,
	})
}

// SimulateDemandaAltaExterna simula un evento de demanda alta desde un sistema externo
func (h *ExternalSimulatorHandler) SimulateDemandaAltaExterna(c *gin.Context) {
	var request struct {
		ProductoID          string  `json:"producto_id" binding:"required"`
		NombreProducto      string  `json:"nombre_producto" binding:"required"`
		DemandaPronosticada int     `json:"demanda_pronosticada" binding:"required"`
		StockActual         int     `json:"stock_actual" binding:"required"`
		CantidadRequerida   int     `json:"cantidad_requerida" binding:"required"`
		ConfianzaPronostico float64 `json:"confianza_pronostico" binding:"required"`
		PeriodoPronostico   string  `json:"periodo_pronostico"`
		Prioridad           string  `json:"prioridad"`
		Source              string  `json:"source"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Valores por defecto
	if request.PeriodoPronostico == "" {
		request.PeriodoPronostico = "30 días"
	}
	if request.Prioridad == "" {
		request.Prioridad = "MEDIA"
	}
	if request.Source == "" {
		request.Source = "Sistema de Pronóstico Externo"
	}

	event := events.DemandaAltaExternaEvent{
		EventID:    uuid.New().String(),
		EventType:  events.EventTypeDemandaAltaExterna,
		ProductoID: request.ProductoID,
		Timestamp:  time.Now(),
		Source:     request.Source,
		Data: struct {
			NombreProducto      string  `json:"nombre_producto"`
			DemandaPronosticada int     `json:"demanda_pronosticada"`
			StockActual         int     `json:"stock_actual"`
			CantidadRequerida   int     `json:"cantidad_requerida"`
			ConfianzaPronostico float64 `json:"confianza_pronostico"`
			PeriodoPronostico   string  `json:"periodo_pronostico"`
			Prioridad           string  `json:"prioridad"`
		}{
			NombreProducto:      request.NombreProducto,
			DemandaPronosticada: request.DemandaPronosticada,
			StockActual:         request.StockActual,
			CantidadRequerida:   request.CantidadRequerida,
			ConfianzaPronostico: request.ConfianzaPronostico,
			PeriodoPronostico:   request.PeriodoPronostico,
			Prioridad:           request.Prioridad,
		},
	}

	// Publicar evento
	err := h.eventBus.Publish(events.TopicExternalEvents, &event)
	if err != nil {
		h.log.WithError(err).Error("Failed to publish demanda alta externa event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
		return
	}

	h.log.WithFields(logrus.Fields{
		"producto_id":          request.ProductoID,
		"demanda_pronosticada": request.DemandaPronosticada,
		"confianza":            request.ConfianzaPronostico,
		"source":               request.Source,
	}).Info("Demanda alta externa event published")

	c.JSON(http.StatusOK, gin.H{
		"message":    "Demanda alta externa event published successfully",
		"event_id":   event.EventID,
		"event_type": event.EventType,
	})
}

// SimulateLoteDanadoExterno simula un evento de lote dañado desde un sistema externo
func (h *ExternalSimulatorHandler) SimulateLoteDanadoExterno(c *gin.Context) {
	var request struct {
		ProductoID            string  `json:"producto_id" binding:"required"`
		NombreProducto        string  `json:"nombre_producto" binding:"required"`
		LoteID                string  `json:"lote_id" binding:"required"`
		CantidadDanada        int     `json:"cantidad_danada" binding:"required"`
		TemperaturaRegistrada float64 `json:"temperatura_registrada" binding:"required"`
		TemperaturaRequerida  float64 `json:"temperatura_requerida" binding:"required"`
		MotivoDanio           string  `json:"motivo_danio" binding:"required"`
		Urgencia              string  `json:"urgencia"`
		Source                string  `json:"source"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Valores por defecto
	if request.Urgencia == "" {
		request.Urgencia = "ALTA"
	}
	if request.Source == "" {
		request.Source = "Sistema de Monitoreo de Temperatura"
	}

	event := events.LoteDanadoExternoEvent{
		EventID:    uuid.New().String(),
		EventType:  events.EventTypeLoteDanadoExterno,
		ProductoID: request.ProductoID,
		Timestamp:  time.Now(),
		Source:     request.Source,
		Data: struct {
			NombreProducto        string  `json:"nombre_producto"`
			LoteID                string  `json:"lote_id"`
			CantidadDanada        int     `json:"cantidad_danada"`
			TemperaturaRegistrada float64 `json:"temperatura_registrada"`
			TemperaturaRequerida  float64 `json:"temperatura_requerida"`
			MotivoDanio           string  `json:"motivo_danio"`
			Urgencia              string  `json:"urgencia"`
		}{
			NombreProducto:        request.NombreProducto,
			LoteID:                request.LoteID,
			CantidadDanada:        request.CantidadDanada,
			TemperaturaRegistrada: request.TemperaturaRegistrada,
			TemperaturaRequerida:  request.TemperaturaRequerida,
			MotivoDanio:           request.MotivoDanio,
			Urgencia:              request.Urgencia,
		},
	}

	// Publicar evento
	err := h.eventBus.Publish(events.TopicExternalEvents, &event)
	if err != nil {
		h.log.WithError(err).Error("Failed to publish lote danado externo event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
		return
	}

	h.log.WithFields(logrus.Fields{
		"producto_id":     request.ProductoID,
		"lote_id":         request.LoteID,
		"cantidad_danada": request.CantidadDanada,
		"motivo":          request.MotivoDanio,
		"source":          request.Source,
	}).Info("Lote danado externo event published")

	c.JSON(http.StatusOK, gin.H{
		"message":    "Lote danado externo event published successfully",
		"event_id":   event.EventID,
		"event_type": event.EventType,
	})
}

// SimulateAlertaInventarioExterna simula una alerta de inventario desde un sistema externo
func (h *ExternalSimulatorHandler) SimulateAlertaInventarioExterna(c *gin.Context) {
	var request struct {
		ProductoID        string `json:"producto_id" binding:"required"`
		NombreProducto    string `json:"nombre_producto" binding:"required"`
		TipoAlerta        string `json:"tipo_alerta" binding:"required"`
		Descripcion       string `json:"descripcion" binding:"required"`
		StockActual       int    `json:"stock_actual" binding:"required"`
		StockMinimo       int    `json:"stock_minimo" binding:"required"`
		CantidadRequerida int    `json:"cantidad_requerida" binding:"required"`
		Prioridad         string `json:"prioridad"`
		FechaVencimiento  string `json:"fecha_vencimiento"`
		Source            string `json:"source"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Valores por defecto
	if request.Prioridad == "" {
		request.Prioridad = "MEDIA"
	}
	if request.Source == "" {
		request.Source = "Sistema de Gestión de Inventario"
	}

	event := events.AlertaInventarioExternaEvent{
		EventID:    uuid.New().String(),
		EventType:  events.EventTypeAlertaInventarioExterna,
		ProductoID: request.ProductoID,
		Timestamp:  time.Now(),
		Source:     request.Source,
		Data: struct {
			NombreProducto    string `json:"nombre_producto"`
			TipoAlerta        string `json:"tipo_alerta"`
			Descripcion       string `json:"descripcion"`
			StockActual       int    `json:"stock_actual"`
			StockMinimo       int    `json:"stock_minimo"`
			CantidadRequerida int    `json:"cantidad_requerida"`
			Prioridad         string `json:"prioridad"`
			FechaVencimiento  string `json:"fecha_vencimiento,omitempty"`
		}{
			NombreProducto:    request.NombreProducto,
			TipoAlerta:        request.TipoAlerta,
			Descripcion:       request.Descripcion,
			StockActual:       request.StockActual,
			StockMinimo:       request.StockMinimo,
			CantidadRequerida: request.CantidadRequerida,
			Prioridad:         request.Prioridad,
			FechaVencimiento:  request.FechaVencimiento,
		},
	}

	// Publicar evento
	err := h.eventBus.Publish(events.TopicExternalEvents, &event)
	if err != nil {
		h.log.WithError(err).Error("Failed to publish alerta inventario externa event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
		return
	}

	h.log.WithFields(logrus.Fields{
		"producto_id":        request.ProductoID,
		"tipo_alerta":        request.TipoAlerta,
		"stock_actual":       request.StockActual,
		"cantidad_requerida": request.CantidadRequerida,
		"source":             request.Source,
	}).Info("Alerta inventario externa event published")

	c.JSON(http.StatusOK, gin.H{
		"message":    "Alerta inventario externa event published successfully",
		"event_id":   event.EventID,
		"event_type": event.EventType,
	})
}

// GetExternalEventTypes devuelve los tipos de eventos externos disponibles
func (h *ExternalSimulatorHandler) GetExternalEventTypes(c *gin.Context) {
	eventTypes := []map[string]string{
		{
			"type":        events.EventTypeStockBajoExterno,
			"description": "Stock bajo detectado por sistema externo",
			"endpoint":    "/api/v1/external/simulate/stock-bajo",
		},
		{
			"type":        events.EventTypeDemandaAltaExterna,
			"description": "Demanda alta pronosticada por sistema externo",
			"endpoint":    "/api/v1/external/simulate/demanda-alta",
		},
		{
			"type":        events.EventTypeLoteDanadoExterno,
			"description": "Lote dañado detectado por sistema externo",
			"endpoint":    "/api/v1/external/simulate/lote-danado",
		},
		{
			"type":        events.EventTypeAlertaInventarioExterna,
			"description": "Alerta de inventario desde sistema externo",
			"endpoint":    "/api/v1/external/simulate/alerta-inventario",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Available external event types",
		"event_types": eventTypes,
	})
}
