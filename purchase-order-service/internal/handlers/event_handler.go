package handlers

import (
	"encoding/json"
	"mediplus/purchase-order-service/internal/events"
	"mediplus/purchase-order-service/internal/service"

	"github.com/sirupsen/logrus"
)

// EventHandler maneja los eventos recibidos del event bus
type EventHandler struct {
	orderService service.OrderService
	log          *logrus.Logger
}

// NewEventHandler crea una nueva instancia de EventHandler
func NewEventHandler(orderService service.OrderService, log *logrus.Logger) *EventHandler {
	return &EventHandler{
		orderService: orderService,
		log:          log,
	}
}

// HandleStockBajoEvent maneja eventos de stock bajo
func (h *EventHandler) HandleStockBajoEvent(eventData []byte) error {
	h.log.Info("Received StockBajo event")

	var stockEvent events.StockBajoEvent
	if err := json.Unmarshal(eventData, &stockEvent); err != nil {
		h.log.Errorf("Error unmarshaling StockBajo event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id":           stockEvent.EventID,
		"producto_id":        stockEvent.ProductoID,
		"stock_actual":       stockEvent.Data.StockActual,
		"punto_reorden":      stockEvent.Data.PuntoReorden,
		"cantidad_requerida": stockEvent.Data.CantidadRequerida,
	}).Info("Processing StockBajo event")

	// Procesar el evento de stock bajo y crear orden automáticamente
	err := h.orderService.ProcessStockLowEvent(stockEvent.ProductoID)
	if err != nil {
		h.log.Errorf("Error processing stock low event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id":    stockEvent.EventID,
		"producto_id": stockEvent.ProductoID,
	}).Info("Successfully processed StockBajo event and created order")

	return nil
}

// HandleLoteDanadoEvent maneja eventos de lote dañado
func (h *EventHandler) HandleLoteDanadoEvent(eventData []byte) error {
	h.log.Info("Received LoteDanado event")

	var loteEvent events.LoteDanadoEvent
	if err := json.Unmarshal(eventData, &loteEvent); err != nil {
		h.log.Errorf("Error unmarshaling LoteDanado event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id":        loteEvent.EventID,
		"producto_id":     loteEvent.ProductoID,
		"lote_id":         loteEvent.Data.LoteID,
		"cantidad_danada": loteEvent.Data.CantidadDanada,
	}).Info("Processing LoteDanado event")

	// Procesar el evento de lote dañado y crear orden automáticamente
	err := h.orderService.ProcessLoteDanadoEvent(
		loteEvent.ProductoID,
		loteEvent.Data.LoteID,
		loteEvent.Data.CantidadDanada,
		loteEvent.Data.TemperaturaRegistrada,
	)
	if err != nil {
		h.log.Errorf("Error processing damaged batch event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id":    loteEvent.EventID,
		"producto_id": loteEvent.ProductoID,
		"lote_id":     loteEvent.Data.LoteID,
	}).Info("Successfully processed LoteDanado event and created order")

	return nil
}

// HandlePronosticoDemandaAltaEvent maneja eventos de pronóstico de alta demanda
func (h *EventHandler) HandlePronosticoDemandaAltaEvent(eventData []byte) error {
	h.log.Info("Received PronosticoDemandaAlta event")

	var pronosticoEvent events.PronosticoDemandaAltaEvent
	if err := json.Unmarshal(eventData, &pronosticoEvent); err != nil {
		h.log.Errorf("Error unmarshaling PronosticoDemandaAlta event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id":             pronosticoEvent.EventID,
		"producto_id":          pronosticoEvent.ProductoID,
		"demanda_pronosticada": pronosticoEvent.Data.DemandaPronosticada,
		"stock_actual":         pronosticoEvent.Data.StockActual,
	}).Info("Processing PronosticoDemandaAlta event")

	// Procesar el evento de pronóstico de alta demanda y crear orden automáticamente
	err := h.orderService.ProcessPronosticoDemandaAltaEvent(
		pronosticoEvent.ProductoID,
		pronosticoEvent.Data.DemandaPronosticada,
	)
	if err != nil {
		h.log.Errorf("Error processing high demand forecast event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id":    pronosticoEvent.EventID,
		"producto_id": pronosticoEvent.ProductoID,
	}).Info("Successfully processed PronosticoDemandaAlta event and created order")

	return nil
}
