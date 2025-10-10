package handlers

import (
	"mediplus/supplier-service/internal/models"
	"mediplus/supplier-service/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SupplierHandler maneja las peticiones HTTP para proveedores
type SupplierHandler struct {
	service service.SupplierService
	log     *logrus.Logger
}

// NewSupplierHandler crea una nueva instancia de SupplierHandler
func NewSupplierHandler(service service.SupplierService, log *logrus.Logger) *SupplierHandler {
	return &SupplierHandler{
		service: service,
		log:     log,
	}
}

// CreateSupplierRequest representa la petición para crear un proveedor
type CreateSupplierRequest struct {
	NombreLegal          string                     `json:"nombre_legal" binding:"required"`
	RazonSocial          string                     `json:"razon_social" binding:"required"`
	IdentificacionFiscal string                     `json:"identificacion_fiscal" binding:"required"`
	Contactos            []models.ContactoProveedor `json:"contactos"`
	ProductosOfrecidos   []models.ProductoOfrecido  `json:"productos_ofrecidos"`
	Certificaciones      []models.Certificacion     `json:"certificaciones"`
	CapacidadLogistica   *models.CapacidadLogistica `json:"capacidad_logistica"`
}

// UpdateSupplierRequest representa la petición para actualizar un proveedor
type UpdateSupplierRequest struct {
	NombreLegal          string                     `json:"nombre_legal"`
	RazonSocial          string                     `json:"razon_social"`
	IdentificacionFiscal string                     `json:"identificacion_fiscal"`
	Contactos            []models.ContactoProveedor `json:"contactos"`
	ProductosOfrecidos   []models.ProductoOfrecido  `json:"productos_ofrecidos"`
	Certificaciones      []models.Certificacion     `json:"certificaciones"`
	CapacidadLogistica   *models.CapacidadLogistica `json:"capacidad_logistica"`
}

// EvaluateSupplierRequest representa la petición para evaluar un proveedor
type EvaluateSupplierRequest struct {
	ScoreGeneral         float64 `json:"score_general" binding:"required"`
	CumplimientoPlazos   float64 `json:"cumplimiento_plazos" binding:"required"`
	CalidadProductos     float64 `json:"calidad_productos" binding:"required"`
	RespuestaEmergencias float64 `json:"respuesta_emergencias" binding:"required"`
}

// SuspendSupplierRequest representa la petición para suspender un proveedor
type SuspendSupplierRequest struct {
	Motivo string `json:"motivo" binding:"required"`
}

// CreateSupplier crea un nuevo proveedor
func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	var req CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Crear el proveedor
	proveedor := models.NewProveedor(req.NombreLegal, req.RazonSocial, req.IdentificacionFiscal)
	proveedor.Contactos = req.Contactos
	proveedor.ProductosOfrecidos = req.ProductosOfrecidos
	proveedor.Certificaciones = req.Certificaciones
	proveedor.CapacidadLogistica = req.CapacidadLogistica

	err := h.service.CreateSupplier(proveedor)
	if err != nil {
		h.log.Errorf("Error creating supplier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating supplier"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Supplier created successfully",
		"data":    proveedor,
	})
}

// GetSupplier obtiene un proveedor por ID
func (h *SupplierHandler) GetSupplier(c *gin.Context) {
	proveedorID := c.Param("id")
	if proveedorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Supplier ID is required"})
		return
	}

	proveedor, err := h.service.GetSupplier(proveedorID)
	if err != nil {
		h.log.Errorf("Error getting supplier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting supplier"})
		return
	}

	if proveedor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": proveedor})
}

// UpdateSupplier actualiza un proveedor
func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	proveedorID := c.Param("id")
	if proveedorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Supplier ID is required"})
		return
	}

	var req UpdateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Obtener el proveedor actual
	proveedor, err := h.service.GetSupplier(proveedorID)
	if err != nil {
		h.log.Errorf("Error getting supplier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting supplier"})
		return
	}

	if proveedor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
		return
	}

	// Actualizar campos
	if req.NombreLegal != "" {
		proveedor.NombreLegal = req.NombreLegal
	}
	if req.RazonSocial != "" {
		proveedor.RazonSocial = req.RazonSocial
	}
	if req.IdentificacionFiscal != "" {
		proveedor.IdentificacionFiscal = req.IdentificacionFiscal
	}
	if req.Contactos != nil {
		proveedor.Contactos = req.Contactos
	}
	if req.ProductosOfrecidos != nil {
		proveedor.ProductosOfrecidos = req.ProductosOfrecidos
	}
	if req.Certificaciones != nil {
		proveedor.Certificaciones = req.Certificaciones
	}
	if req.CapacidadLogistica != nil {
		proveedor.CapacidadLogistica = req.CapacidadLogistica
	}

	err = h.service.UpdateSupplier(proveedor)
	if err != nil {
		h.log.Errorf("Error updating supplier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating supplier"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Supplier updated successfully",
		"data":    proveedor,
	})
}

// DeleteSupplier elimina un proveedor
func (h *SupplierHandler) DeleteSupplier(c *gin.Context) {
	proveedorID := c.Param("id")
	if proveedorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Supplier ID is required"})
		return
	}

	err := h.service.DeleteSupplier(proveedorID)
	if err != nil {
		h.log.Errorf("Error deleting supplier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting supplier"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Supplier deleted successfully"})
}

// ListSuppliers lista todos los proveedores
func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	// Obtener parámetros de consulta
	estado := c.Query("estado")
	certificacion := c.Query("certificacion")
	cadenaFrio := c.Query("cadena_frio")

	var proveedores []*models.Proveedor
	var err error

	if estado != "" {
		// Listar por estado
		estadoProveedor := models.EstadoProveedor(estado)
		proveedores, err = h.service.ListSuppliersByEstado(estadoProveedor)
	} else if certificacion != "" {
		// Listar por certificación
		proveedores, err = h.service.GetSuppliersByCertification(certificacion)
	} else if cadenaFrio == "true" {
		// Listar por capacidad de cadena de frío
		proveedores, err = h.service.GetSuppliersWithColdChain()
	} else {
		// Listar todos
		proveedores, err = h.service.ListSuppliers()
	}

	if err != nil {
		h.log.Errorf("Error listing suppliers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listing suppliers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": proveedores})
}

// EvaluateSupplier evalúa un proveedor
func (h *SupplierHandler) EvaluateSupplier(c *gin.Context) {
	proveedorID := c.Param("id")
	if proveedorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Supplier ID is required"})
		return
	}

	var req EvaluateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Crear evaluación
	evaluacion := &models.EvaluacionRendimiento{
		ScoreGeneral:             req.ScoreGeneral,
		CumplimientoPlazos:       req.CumplimientoPlazos,
		CalidadProductos:         req.CalidadProductos,
		RespuestaEmergencias:     req.RespuestaEmergencias,
		FechaUltimaActualizacion: time.Now(),
	}

	err := h.service.EvaluateSupplier(proveedorID, evaluacion)
	if err != nil {
		h.log.Errorf("Error evaluating supplier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error evaluating supplier"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Supplier evaluated successfully",
		"data":    evaluacion,
	})
}

// SuspendSupplier suspende un proveedor
func (h *SupplierHandler) SuspendSupplier(c *gin.Context) {
	proveedorID := c.Param("id")
	if proveedorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Supplier ID is required"})
		return
	}

	var req SuspendSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorf("Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.SuspendSupplier(proveedorID, req.Motivo)
	if err != nil {
		h.log.Errorf("Error suspending supplier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error suspending supplier"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Supplier suspended successfully"})
}

// ActivateSupplier activa un proveedor
func (h *SupplierHandler) ActivateSupplier(c *gin.Context) {
	proveedorID := c.Param("id")
	if proveedorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Supplier ID is required"})
		return
	}

	err := h.service.ActivateSupplier(proveedorID)
	if err != nil {
		h.log.Errorf("Error activating supplier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error activating supplier"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Supplier activated successfully"})
}
