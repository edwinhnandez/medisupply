package service

import (
	"mediplus/supplier-service/internal/events"
	"mediplus/supplier-service/internal/models"
	"mediplus/supplier-service/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// SupplierService define la interfaz para el servicio de proveedores
type SupplierService interface {
	CreateSupplier(proveedor *models.Proveedor) error
	GetSupplier(proveedorID string) (*models.Proveedor, error)
	UpdateSupplier(proveedor *models.Proveedor) error
	DeleteSupplier(proveedorID string) error
	ListSuppliers() ([]*models.Proveedor, error)
	EvaluateSupplier(proveedorID string, evaluacion *models.EvaluacionRendimiento) error
	SuspendSupplier(proveedorID string, motivo string) error
	ActivateSupplier(proveedorID string) error
	GetSuppliersByCertification(tipoCertificacion string) ([]*models.Proveedor, error)
	GetSuppliersWithColdChain() ([]*models.Proveedor, error)
	ListSuppliersByEstado(estado models.EstadoProveedor) ([]*models.Proveedor, error)
	CheckExpiringCertifications() error
	ProcessOrderGeneratedEvent(orderEvent *events.OrdenCompraGeneradaEvent) error
	ProcessOrderConfirmedEvent(orderEvent *events.OrdenCompraConfirmadaEvent) error
	ProcessOrderReceivedEvent(orderEvent *events.OrdenCompraRecibidaEvent) error
}

// supplierService implementa SupplierService
type supplierService struct {
	supplierRepo repository.SupplierRepository
	auditRepo    repository.AuditRepository
	eventBus     events.EventBus
	log          *logrus.Logger
}

// NewSupplierService crea una nueva instancia de SupplierService
func NewSupplierService(
	supplierRepo repository.SupplierRepository,
	auditRepo repository.AuditRepository,
	eventBus events.EventBus,
	log *logrus.Logger,
) SupplierService {
	return &supplierService{
		supplierRepo: supplierRepo,
		auditRepo:    auditRepo,
		eventBus:     eventBus,
		log:          log,
	}
}

// CreateSupplier crea un nuevo proveedor
func (s *supplierService) CreateSupplier(proveedor *models.Proveedor) error {
	// Crear el proveedor
	err := s.supplierRepo.Create(proveedor)
	if err != nil {
		s.log.Errorf("Error creating supplier: %v", err)
		return err
	}

	// Crear traza de auditoría
	traza := &models.AuditoriaTraza{
		TrazaID:       uuid.New().String(),
		ProveedorID:   proveedor.ProveedorID,
		TipoCambio:    "CREACION",
		Descripcion:   "Proveedor creado",
		ValorAnterior: "",
		ValorNuevo:    proveedor.NombreLegal,
		UsuarioID:     "system",
		FechaCambio:   time.Now(),
		IPAddress:     "127.0.0.1",
	}

	err = s.auditRepo.CreateTraza(traza)
	if err != nil {
		s.log.Errorf("Error creating audit trace: %v", err)
		// No retornamos error aquí para no afectar la creación del proveedor
	}

	// Emitir evento
	event := &events.ProveedorCalificadoEvent{
		EventID:     uuid.New().String(),
		EventType:   events.EventTypeProveedorCalificado,
		ProveedorID: proveedor.ProveedorID,
		Timestamp:   time.Now(),
	}

	event.Data.NombreLegal = proveedor.NombreLegal
	event.Data.RazonSocial = proveedor.RazonSocial
	if proveedor.EvaluacionRendimiento != nil {
		event.Data.ScoreGeneral = proveedor.EvaluacionRendimiento.ScoreGeneral
	}
	if proveedor.CapacidadLogistica != nil {
		event.Data.CapacidadCadenaFrio = proveedor.CapacidadLogistica.CapacidadCadenaFrio
	}

	// Obtener certificaciones
	for _, cert := range proveedor.Certificaciones {
		event.Data.Certificaciones = append(event.Data.Certificaciones, cert.TipoCertificacion)
	}

	err = s.eventBus.Publish(events.TopicProveedorEvents, event)
	if err != nil {
		s.log.Errorf("Error publishing event: %v", err)
		// No retornamos error aquí para no afectar la creación del proveedor
	}

	return nil
}

// GetSupplier obtiene un proveedor por su ID
func (s *supplierService) GetSupplier(proveedorID string) (*models.Proveedor, error) {
	return s.supplierRepo.GetByID(proveedorID)
}

// UpdateSupplier actualiza un proveedor
func (s *supplierService) UpdateSupplier(proveedor *models.Proveedor) error {
	// Obtener el proveedor actual para comparar cambios
	proveedorActual, err := s.supplierRepo.GetByID(proveedor.ProveedorID)
	if err != nil {
		return err
	}

	if proveedorActual == nil {
		return nil // Proveedor no encontrado
	}

	// Actualizar el proveedor
	err = s.supplierRepo.Update(proveedor)
	if err != nil {
		s.log.Errorf("Error updating supplier: %v", err)
		return err
	}

	// Crear traza de auditoría
	traza := &models.AuditoriaTraza{
		TrazaID:       uuid.New().String(),
		ProveedorID:   proveedor.ProveedorID,
		TipoCambio:    "ACTUALIZACION",
		Descripcion:   "Proveedor actualizado",
		ValorAnterior: proveedorActual.NombreLegal,
		ValorNuevo:    proveedor.NombreLegal,
		UsuarioID:     "system",
		FechaCambio:   time.Now(),
		IPAddress:     "127.0.0.1",
	}

	err = s.auditRepo.CreateTraza(traza)
	if err != nil {
		s.log.Errorf("Error creating audit trace: %v", err)
	}

	return nil
}

// DeleteSupplier elimina un proveedor
func (s *supplierService) DeleteSupplier(proveedorID string) error {
	// Obtener el proveedor antes de eliminarlo
	proveedor, err := s.supplierRepo.GetByID(proveedorID)
	if err != nil {
		return err
	}

	if proveedor == nil {
		return nil // Proveedor no encontrado
	}

	// Eliminar el proveedor
	err = s.supplierRepo.Delete(proveedorID)
	if err != nil {
		s.log.Errorf("Error deleting supplier: %v", err)
		return err
	}

	// Crear traza de auditoría
	traza := &models.AuditoriaTraza{
		TrazaID:       uuid.New().String(),
		ProveedorID:   proveedorID,
		TipoCambio:    "ELIMINACION",
		Descripcion:   "Proveedor eliminado",
		ValorAnterior: proveedor.NombreLegal,
		ValorNuevo:    "",
		UsuarioID:     "system",
		FechaCambio:   time.Now(),
		IPAddress:     "127.0.0.1",
	}

	err = s.auditRepo.CreateTraza(traza)
	if err != nil {
		s.log.Errorf("Error creating audit trace: %v", err)
	}

	return nil
}

// ListSuppliers lista todos los proveedores
func (s *supplierService) ListSuppliers() ([]*models.Proveedor, error) {
	return s.supplierRepo.ListAll()
}

// EvaluateSupplier evalúa un proveedor
func (s *supplierService) EvaluateSupplier(proveedorID string, evaluacion *models.EvaluacionRendimiento) error {
	// Obtener el proveedor actual
	proveedor, err := s.supplierRepo.GetByID(proveedorID)
	if err != nil {
		return err
	}

	if proveedor == nil {
		return nil // Proveedor no encontrado
	}

	// Guardar score anterior para el evento
	scoreAnterior := float64(0)
	if proveedor.EvaluacionRendimiento != nil {
		scoreAnterior = proveedor.EvaluacionRendimiento.ScoreGeneral
	}

	// Actualizar la evaluación
	proveedor.EvaluacionRendimiento = evaluacion
	proveedor.FechaUltimaEvaluacion = time.Now()

	// Actualizar el proveedor
	err = s.supplierRepo.Update(proveedor)
	if err != nil {
		s.log.Errorf("Error updating supplier evaluation: %v", err)
		return err
	}

	// Crear traza de auditoría
	traza := &models.AuditoriaTraza{
		TrazaID:       uuid.New().String(),
		ProveedorID:   proveedorID,
		TipoCambio:    "EVALUACION",
		Descripcion:   "Evaluación de rendimiento actualizada",
		ValorAnterior: string(rune(scoreAnterior)),
		ValorNuevo:    string(rune(evaluacion.ScoreGeneral)),
		UsuarioID:     "system",
		FechaCambio:   time.Now(),
		IPAddress:     "127.0.0.1",
	}

	err = s.auditRepo.CreateTraza(traza)
	if err != nil {
		s.log.Errorf("Error creating audit trace: %v", err)
	}

	// Emitir evento de evaluación actualizada
	event := &events.EvaluacionActualizadaEvent{
		EventID:     uuid.New().String(),
		EventType:   events.EventTypeEvaluacionActualizada,
		ProveedorID: proveedorID,
		Timestamp:   time.Now(),
	}

	event.Data.ScoreAnterior = scoreAnterior
	event.Data.ScoreNuevo = evaluacion.ScoreGeneral
	event.Data.CumplimientoPlazos = evaluacion.CumplimientoPlazos
	event.Data.CalidadProductos = evaluacion.CalidadProductos
	event.Data.RespuestaEmergencias = evaluacion.RespuestaEmergencias

	err = s.eventBus.Publish(events.TopicProveedorEvents, event)
	if err != nil {
		s.log.Errorf("Error publishing evaluation event: %v", err)
	}

	return nil
}

// SuspendSupplier suspende un proveedor
func (s *supplierService) SuspendSupplier(proveedorID string, motivo string) error {
	// Obtener el proveedor actual
	proveedor, err := s.supplierRepo.GetByID(proveedorID)
	if err != nil {
		return err
	}

	if proveedor == nil {
		return nil // Proveedor no encontrado
	}

	// Actualizar estado
	proveedor.EstadoProveedor = models.EstadoSuspendido

	// Actualizar el proveedor
	err = s.supplierRepo.Update(proveedor)
	if err != nil {
		s.log.Errorf("Error suspending supplier: %v", err)
		return err
	}

	// Crear traza de auditoría
	traza := &models.AuditoriaTraza{
		TrazaID:       uuid.New().String(),
		ProveedorID:   proveedorID,
		TipoCambio:    "SUSPENSION",
		Descripcion:   "Proveedor suspendido: " + motivo,
		ValorAnterior: string(models.EstadoActivo),
		ValorNuevo:    string(models.EstadoSuspendido),
		UsuarioID:     "system",
		FechaCambio:   time.Now(),
		IPAddress:     "127.0.0.1",
	}

	err = s.auditRepo.CreateTraza(traza)
	if err != nil {
		s.log.Errorf("Error creating audit trace: %v", err)
	}

	// Emitir evento de suspensión
	event := &events.ProveedorSuspendidoEvent{
		EventID:     uuid.New().String(),
		EventType:   events.EventTypeProveedorSuspendido,
		ProveedorID: proveedorID,
		Timestamp:   time.Now(),
	}

	event.Data.MotivoSuspension = motivo
	event.Data.FechaSuspension = time.Now()

	err = s.eventBus.Publish(events.TopicProveedorEvents, event)
	if err != nil {
		s.log.Errorf("Error publishing suspension event: %v", err)
	}

	return nil
}

// ActivateSupplier activa un proveedor
func (s *supplierService) ActivateSupplier(proveedorID string) error {
	// Obtener el proveedor actual
	proveedor, err := s.supplierRepo.GetByID(proveedorID)
	if err != nil {
		return err
	}

	if proveedor == nil {
		return nil // Proveedor no encontrado
	}

	// Actualizar estado
	proveedor.EstadoProveedor = models.EstadoActivo

	// Actualizar el proveedor
	err = s.supplierRepo.Update(proveedor)
	if err != nil {
		s.log.Errorf("Error activating supplier: %v", err)
		return err
	}

	// Crear traza de auditoría
	traza := &models.AuditoriaTraza{
		TrazaID:       uuid.New().String(),
		ProveedorID:   proveedorID,
		TipoCambio:    "ACTIVACION",
		Descripcion:   "Proveedor activado",
		ValorAnterior: string(models.EstadoSuspendido),
		ValorNuevo:    string(models.EstadoActivo),
		UsuarioID:     "system",
		FechaCambio:   time.Now(),
		IPAddress:     "127.0.0.1",
	}

	err = s.auditRepo.CreateTraza(traza)
	if err != nil {
		s.log.Errorf("Error creating audit trace: %v", err)
	}

	// Emitir evento de activación
	event := &events.ProveedorCalificadoEvent{
		EventID:     uuid.New().String(),
		EventType:   events.EventTypeProveedorActivado,
		ProveedorID: proveedorID,
		Timestamp:   time.Now(),
	}

	event.Data.NombreLegal = proveedor.NombreLegal
	event.Data.RazonSocial = proveedor.RazonSocial
	if proveedor.EvaluacionRendimiento != nil {
		event.Data.ScoreGeneral = proveedor.EvaluacionRendimiento.ScoreGeneral
	}
	if proveedor.CapacidadLogistica != nil {
		event.Data.CapacidadCadenaFrio = proveedor.CapacidadLogistica.CapacidadCadenaFrio
	}

	err = s.eventBus.Publish(events.TopicProveedorEvents, event)
	if err != nil {
		s.log.Errorf("Error publishing activation event: %v", err)
	}

	return nil
}

// GetSuppliersByCertification obtiene proveedores por tipo de certificación
func (s *supplierService) GetSuppliersByCertification(tipoCertificacion string) ([]*models.Proveedor, error) {
	return s.supplierRepo.GetByCertificacion(tipoCertificacion)
}

// GetSuppliersWithColdChain obtiene proveedores con capacidad de cadena de frío
func (s *supplierService) GetSuppliersWithColdChain() ([]*models.Proveedor, error) {
	return s.supplierRepo.GetByCapacidadCadenaFrio()
}

// CheckExpiringCertifications verifica certificaciones por vencer
func (s *supplierService) CheckExpiringCertifications() error {
	// Obtener todos los proveedores
	proveedores, err := s.supplierRepo.ListAll()
	if err != nil {
		return err
	}

	// Verificar cada proveedor
	for _, proveedor := range proveedores {
		for _, cert := range proveedor.Certificaciones {
			// Verificar si la certificación está por vencer (30 días)
			if cert.FechaVencimiento.Before(time.Now().AddDate(0, 0, 30)) &&
				cert.FechaVencimiento.After(time.Now()) {

				// Emitir evento de certificación por vencer
				event := &events.CertificacionPorVencerEvent{
					EventID:     uuid.New().String(),
					EventType:   events.EventTypeCertificacionPorVencer,
					ProveedorID: proveedor.ProveedorID,
					Timestamp:   time.Now(),
				}

				event.Data.CertificacionID = cert.NumeroCertificado
				event.Data.TipoCertificacion = cert.TipoCertificacion
				event.Data.FechaVencimiento = cert.FechaVencimiento
				event.Data.DiasRestantes = int(cert.FechaVencimiento.Sub(time.Now()).Hours() / 24)

				err = s.eventBus.Publish(events.TopicNotifications, event)
				if err != nil {
					s.log.Errorf("Error publishing expiring certification event: %v", err)
				}
			}
		}
	}

	return nil
}

// ListSuppliersByEstado lista proveedores por estado
func (s *supplierService) ListSuppliersByEstado(estado models.EstadoProveedor) ([]*models.Proveedor, error) {
	return s.supplierRepo.ListByEstado(estado)
}

// ProcessOrderGeneratedEvent procesa un evento de orden de compra generada
func (s *supplierService) ProcessOrderGeneratedEvent(orderEvent *events.OrdenCompraGeneradaEvent) error {
	s.log.Infof("Processing order generated event for order: %s", orderEvent.OrdenID)

	// Obtener proveedores activos que puedan cumplir con la orden
	proveedores, err := s.supplierRepo.ListByEstado(models.EstadoActivo)
	if err != nil {
		s.log.Errorf("Error getting active suppliers: %v", err)
		return err
	}

	if len(proveedores) == 0 {
		s.log.Warn("No active suppliers found for order")
		return nil
	}

	// Generar solicitud de proveedor
	return s.generateSupplierRequest(orderEvent, proveedores)
}

// ProcessOrderConfirmedEvent procesa un evento de orden de compra confirmada
func (s *supplierService) ProcessOrderConfirmedEvent(orderEvent *events.OrdenCompraConfirmadaEvent) error {
	s.log.Infof("Processing order confirmed event for order: %s", orderEvent.OrdenID)

	// Obtener el proveedor que confirmó la orden
	proveedor, err := s.supplierRepo.GetByID(orderEvent.Data.ProveedorID)
	if err != nil {
		s.log.Errorf("Error getting supplier: %v", err)
		return err
	}

	if proveedor == nil {
		s.log.Warnf("Supplier not found: %s", orderEvent.Data.ProveedorID)
		return nil
	}

	// Crear traza de auditoría
	traza := &models.AuditoriaTraza{
		TrazaID:       uuid.New().String(),
		ProveedorID:   orderEvent.Data.ProveedorID,
		TipoCambio:    "ORDEN_CONFIRMADA",
		Descripcion:   "Orden de compra confirmada: " + orderEvent.Data.NumeroOrden,
		ValorAnterior: "",
		ValorNuevo:    orderEvent.Data.NumeroOrden,
		UsuarioID:     "system",
		FechaCambio:   time.Now(),
		IPAddress:     "127.0.0.1",
	}

	err = s.auditRepo.CreateTraza(traza)
	if err != nil {
		s.log.Errorf("Error creating audit trace: %v", err)
	}

	s.log.WithFields(logrus.Fields{
		"orden_id":     orderEvent.OrdenID,
		"proveedor_id": orderEvent.Data.ProveedorID,
		"numero_orden": orderEvent.Data.NumeroOrden,
	}).Info("Order confirmed event processed successfully")

	return nil
}

// ProcessOrderReceivedEvent procesa un evento de orden de compra recibida
func (s *supplierService) ProcessOrderReceivedEvent(orderEvent *events.OrdenCompraRecibidaEvent) error {
	s.log.Infof("Processing order received event for order: %s", orderEvent.OrdenID)

	// Obtener el proveedor que entregó la orden
	proveedor, err := s.supplierRepo.GetByID(orderEvent.Data.ProveedorID)
	if err != nil {
		s.log.Errorf("Error getting supplier: %v", err)
		return err
	}

	if proveedor == nil {
		s.log.Warnf("Supplier not found: %s", orderEvent.Data.ProveedorID)
		return nil
	}

	// Crear traza de auditoría
	traza := &models.AuditoriaTraza{
		TrazaID:       uuid.New().String(),
		ProveedorID:   orderEvent.Data.ProveedorID,
		TipoCambio:    "ORDEN_RECIBIDA",
		Descripcion:   "Orden de compra recibida: " + orderEvent.Data.NumeroOrden,
		ValorAnterior: "",
		ValorNuevo:    orderEvent.Data.NumeroOrden,
		UsuarioID:     "system",
		FechaCambio:   time.Now(),
		IPAddress:     "127.0.0.1",
	}

	err = s.auditRepo.CreateTraza(traza)
	if err != nil {
		s.log.Errorf("Error creating audit trace: %v", err)
	}

	s.log.WithFields(logrus.Fields{
		"orden_id":     orderEvent.OrdenID,
		"proveedor_id": orderEvent.Data.ProveedorID,
		"numero_orden": orderEvent.Data.NumeroOrden,
	}).Info("Order received event processed successfully")

	return nil
}

// generateSupplierRequest genera una solicitud de proveedor basada en la orden
func (s *supplierService) generateSupplierRequest(orderEvent *events.OrdenCompraGeneradaEvent, proveedores []*models.Proveedor) error {
	s.log.Infof("Generating supplier request for order: %s", orderEvent.OrdenID)

	// Crear evento de solicitud de proveedor
	event := &events.SolicitudProveedorEvent{
		EventID:   uuid.New().String(),
		EventType: events.EventTypeSolicitudProveedor,
		OrdenID:   orderEvent.OrdenID,
		Timestamp: time.Now(),
	}

	// Llenar datos del evento
	event.Data.NumeroOrden = orderEvent.Data.NumeroOrden
	event.Data.Prioridad = orderEvent.Data.Prioridad
	event.Data.MotivoGeneracion = orderEvent.Data.MotivoGeneracion
	event.Data.TotalItems = orderEvent.Data.TotalItems
	event.Data.ValorTotal = orderEvent.Data.ValorTotal

	// Determinar requisitos especiales basados en la prioridad
	event.Data.RequisitosEspeciales = s.determineSpecialRequirements(orderEvent.Data.Prioridad)

	// Agregar información de proveedores disponibles
	event.Data.ProductosRequeridos = s.buildProductRequirements(orderEvent, proveedores)

	// Publicar evento de solicitud de proveedor
	err := s.eventBus.Publish(events.TopicProveedorEvents, event)
	if err != nil {
		s.log.Errorf("Error publishing supplier request event: %v", err)
		return err
	}

	s.log.WithFields(logrus.Fields{
		"event_id":     event.EventID,
		"orden_id":     orderEvent.OrdenID,
		"numero_orden": orderEvent.Data.NumeroOrden,
		"prioridad":    orderEvent.Data.Prioridad,
		"total_items":  orderEvent.Data.TotalItems,
		"valor_total":  orderEvent.Data.ValorTotal,
	}).Info("Supplier request event published successfully")

	return nil
}

// determineSpecialRequirements determina los requisitos especiales basados en la prioridad
func (s *supplierService) determineSpecialRequirements(prioridad string) []string {
	requirements := []string{}

	switch prioridad {
	case "CRITICA":
		requirements = append(requirements, "Entrega urgente", "Capacidad de respuesta 24/7", "Certificaciones médicas vigentes")
	case "ALTA":
		requirements = append(requirements, "Entrega rápida", "Certificaciones médicas vigentes")
	case "MEDIA":
		requirements = append(requirements, "Certificaciones médicas vigentes")
	case "BAJA":
		requirements = append(requirements, "Certificaciones básicas")
	}

	return requirements
}

// buildProductRequirements construye los requisitos de productos basados en la orden
func (s *supplierService) buildProductRequirements(orderEvent *events.OrdenCompraGeneradaEvent, proveedores []*models.Proveedor) []events.ProductoRequerido {
	// Por simplicidad, crear un producto requerido genérico
	// En un escenario real, esto vendría de la información detallada de la orden
	productoRequerido := events.ProductoRequerido{
		ProductoID:           "prod-generic",
		NombreProducto:       "Producto Médico Genérico",
		CantidadRequerida:    orderEvent.Data.TotalItems,
		PrecioUnitario:       orderEvent.Data.ValorTotal / float64(orderEvent.Data.TotalItems),
		TemperaturaRequerida: 2.0, // Temperatura de refrigeración por defecto
		RequiereCadenaFrio:   true,
	}

	return []events.ProductoRequerido{productoRequerido}
}
