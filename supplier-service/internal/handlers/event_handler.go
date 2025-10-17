package handlers

import (
	"encoding/json"
	"mediplus/supplier-service/internal/events"
	"mediplus/supplier-service/internal/service"

	"github.com/sirupsen/logrus"
)

// EventHandler maneja los eventos recibidos del event bus
type EventHandler struct {
	supplierService service.SupplierService
	log             *logrus.Logger
}

// NewEventHandler crea una nueva instancia de EventHandler
func NewEventHandler(supplierService service.SupplierService, log *logrus.Logger) *EventHandler {
	return &EventHandler{
		supplierService: supplierService,
		log:             log,
	}
}

// HandleOrdenCompraGeneradaEvent maneja eventos de orden de compra generada
func (h *EventHandler) HandleOrdenCompraGeneradaEvent(eventData []byte) error {
	h.log.Info("Received OrdenCompraGenerada event")

	var orderEvent events.OrdenCompraGeneradaEvent
	if err := json.Unmarshal(eventData, &orderEvent); err != nil {
		h.log.Errorf("Error unmarshaling OrdenCompraGenerada event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id":     orderEvent.EventID,
		"orden_id":     orderEvent.OrdenID,
		"numero_orden": orderEvent.Data.NumeroOrden,
		"prioridad":    orderEvent.Data.Prioridad,
		"total_items":  orderEvent.Data.TotalItems,
		"valor_total":  orderEvent.Data.ValorTotal,
	}).Info("Processing OrdenCompraGenerada event")

	// Procesar el evento y generar solicitud de proveedor
	err := h.supplierService.ProcessOrderGeneratedEvent(&orderEvent)
	if err != nil {
		h.log.Errorf("Error processing order generated event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id": orderEvent.EventID,
		"orden_id": orderEvent.OrdenID,
	}).Info("Successfully processed OrdenCompraGenerada event and generated supplier request")

	return nil
}

// HandleOrdenCompraConfirmadaEvent maneja eventos de orden de compra confirmada
func (h *EventHandler) HandleOrdenCompraConfirmadaEvent(eventData []byte) error {
	h.log.Info("Received OrdenCompraConfirmada event")

	var orderEvent events.OrdenCompraConfirmadaEvent
	if err := json.Unmarshal(eventData, &orderEvent); err != nil {
		h.log.Errorf("Error unmarshaling OrdenCompraConfirmada event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id":           orderEvent.EventID,
		"orden_id":           orderEvent.OrdenID,
		"numero_orden":       orderEvent.Data.NumeroOrden,
		"proveedor_id":       orderEvent.Data.ProveedorID,
		"fecha_confirmacion": orderEvent.Data.FechaConfirmacion,
	}).Info("Processing OrdenCompraConfirmada event")

	// Procesar el evento de orden confirmada
	err := h.supplierService.ProcessOrderConfirmedEvent(&orderEvent)
	if err != nil {
		h.log.Errorf("Error processing order confirmed event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id": orderEvent.EventID,
		"orden_id": orderEvent.OrdenID,
	}).Info("Successfully processed OrdenCompraConfirmada event")

	return nil
}

// HandleOrdenCompraRecibidaEvent maneja eventos de orden de compra recibida
func (h *EventHandler) HandleOrdenCompraRecibidaEvent(eventData []byte) error {
	h.log.Info("Received OrdenCompraRecibida event")

	var orderEvent events.OrdenCompraRecibidaEvent
	if err := json.Unmarshal(eventData, &orderEvent); err != nil {
		h.log.Errorf("Error unmarshaling OrdenCompraRecibida event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id":        orderEvent.EventID,
		"orden_id":        orderEvent.OrdenID,
		"numero_orden":    orderEvent.Data.NumeroOrden,
		"proveedor_id":    orderEvent.Data.ProveedorID,
		"fecha_recepcion": orderEvent.Data.FechaRecepcion,
	}).Info("Processing OrdenCompraRecibida event")

	// Procesar el evento de orden recibida
	err := h.supplierService.ProcessOrderReceivedEvent(&orderEvent)
	if err != nil {
		h.log.Errorf("Error processing order received event: %v", err)
		return err
	}

	h.log.WithFields(logrus.Fields{
		"event_id": orderEvent.EventID,
		"orden_id": orderEvent.OrdenID,
	}).Info("Successfully processed OrdenCompraRecibida event")

	return nil
}
