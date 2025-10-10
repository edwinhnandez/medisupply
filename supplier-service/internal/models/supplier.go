package models

import (
	"time"

	"github.com/google/uuid"
)

// EstadoProveedor representa el estado del proveedor
type EstadoProveedor string

const (
	EstadoActivo     EstadoProveedor = "ACTIVO"
	EstadoSuspendido EstadoProveedor = "SUSPENDIDO"
	EstadoInactivo   EstadoProveedor = "INACTIVO"
)

// Proveedor representa la entidad raíz del agregado ProveedorCalificado
type Proveedor struct {
	ProveedorID           string                 `json:"proveedor_id" dynamodbav:"proveedor_id"`
	NombreLegal           string                 `json:"nombre_legal" dynamodbav:"nombre_legal"`
	RazonSocial           string                 `json:"razon_social" dynamodbav:"razon_social"`
	IdentificacionFiscal  string                 `json:"identificacion_fiscal" dynamodbav:"identificacion_fiscal"`
	EstadoProveedor       EstadoProveedor        `json:"estado_proveedor" dynamodbav:"estado_proveedor"`
	FechaRegistro         time.Time              `json:"fecha_registro" dynamodbav:"fecha_registro"`
	FechaUltimaEvaluacion time.Time              `json:"fecha_ultima_evaluacion" dynamodbav:"fecha_ultima_evaluacion"`
	Contactos             []ContactoProveedor    `json:"contactos" dynamodbav:"contactos"`
	ProductosOfrecidos    []ProductoOfrecido     `json:"productos_ofrecidos" dynamodbav:"productos_ofrecidos"`
	Certificaciones       []Certificacion        `json:"certificaciones" dynamodbav:"certificaciones"`
	EvaluacionRendimiento *EvaluacionRendimiento `json:"evaluacion_rendimiento" dynamodbav:"evaluacion_rendimiento"`
	CapacidadLogistica    *CapacidadLogistica    `json:"capacidad_logistica" dynamodbav:"capacidad_logistica"`
	CreatedAt             time.Time              `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at" dynamodbav:"updated_at"`
}

// ContactoProveedor representa un contacto del proveedor
type ContactoProveedor struct {
	ContactoID          string `json:"contacto_id" dynamodbav:"contacto_id"`
	Nombre              string `json:"nombre" dynamodbav:"nombre"`
	Email               string `json:"email" dynamodbav:"email"`
	Telefono            string `json:"telefono" dynamodbav:"telefono"`
	Cargo               string `json:"cargo" dynamodbav:"cargo"`
	EsContactoPrincipal bool   `json:"es_contacto_principal" dynamodbav:"es_contacto_principal"`
}

// ProductoOfrecido representa un producto ofrecido por el proveedor
type ProductoOfrecido struct {
	ProductoOfrecidoID   string               `json:"producto_ofrecido_id" dynamodbav:"producto_ofrecido_id"`
	ProductoID           string               `json:"producto_id" dynamodbav:"producto_id"`
	CodigoProveedor      string               `json:"codigo_proveedor" dynamodbav:"codigo_proveedor"`
	PrecioBase           float64              `json:"precio_base" dynamodbav:"precio_base"`
	Moneda               string               `json:"moneda" dynamodbav:"moneda"`
	EstadoDisponibilidad EstadoDisponibilidad `json:"estado_disponibilidad" dynamodbav:"estado_disponibilidad"`
}

// EstadoDisponibilidad representa el estado de disponibilidad del producto
type EstadoDisponibilidad string

const (
	EstadoDisponible   EstadoDisponibilidad = "DISPONIBLE"
	EstadoNoDisponible EstadoDisponibilidad = "NO_DISPONIBLE"
	EstadoAgotado      EstadoDisponibilidad = "AGOTADO"
)

// Certificacion representa una certificación del proveedor
type Certificacion struct {
	TipoCertificacion string              `json:"tipo_certificacion" dynamodbav:"tipo_certificacion"`
	NumeroCertificado string              `json:"numero_certificado" dynamodbav:"numero_certificado"`
	FechaEmision      time.Time           `json:"fecha_emision" dynamodbav:"fecha_emision"`
	FechaVencimiento  time.Time           `json:"fecha_vencimiento" dynamodbav:"fecha_vencimiento"`
	AutoridadEmisora  string              `json:"autoridad_emisora" dynamodbav:"autoridad_emisora"`
	Estado            EstadoCertificacion `json:"estado" dynamodbav:"estado"`
}

// EstadoCertificacion representa el estado de la certificación
type EstadoCertificacion string

const (
	EstadoCertificacionActiva    EstadoCertificacion = "ACTIVA"
	EstadoCertificacionVencida   EstadoCertificacion = "VENCIDA"
	EstadoCertificacionPorVencer EstadoCertificacion = "POR_VENCER"
)

// EvaluacionRendimiento representa la evaluación de rendimiento del proveedor
type EvaluacionRendimiento struct {
	ScoreGeneral             float64   `json:"score_general" dynamodbav:"score_general"`
	CumplimientoPlazos       float64   `json:"cumplimiento_plazos" dynamodbav:"cumplimiento_plazos"`
	CalidadProductos         float64   `json:"calidad_productos" dynamodbav:"calidad_productos"`
	RespuestaEmergencias     float64   `json:"respuesta_emergencias" dynamodbav:"respuesta_emergencias"`
	FechaUltimaActualizacion time.Time `json:"fecha_ultima_actualizacion" dynamodbav:"fecha_ultima_actualizacion"`
}

// CapacidadLogistica representa la capacidad logística del proveedor
type CapacidadLogistica struct {
	CapacidadCadenaFrio     bool    `json:"capacidad_cadena_frio" dynamodbav:"capacidad_cadena_frio"`
	TemperaturaMinima       float64 `json:"temperatura_minima" dynamodbav:"temperatura_minima"`
	TemperaturaMaxima       float64 `json:"temperatura_maxima" dynamodbav:"temperatura_maxima"`
	CapacidadAlmacenamiento int     `json:"capacidad_almacenamiento" dynamodbav:"capacidad_almacenamiento"`
	TiempoEntregaPromedio   int     `json:"tiempo_entrega_promedio" dynamodbav:"tiempo_entrega_promedio"`
	ZonasCobertura          string  `json:"zonas_cobertura" dynamodbav:"zonas_cobertura"`
}

// AuditoriaTraza representa una entrada de auditoría
type AuditoriaTraza struct {
	TrazaID       string    `json:"traza_id" dynamodbav:"traza_id"`
	ProveedorID   string    `json:"proveedor_id" dynamodbav:"proveedor_id"`
	TipoCambio    string    `json:"tipo_cambio" dynamodbav:"tipo_cambio"`
	Descripcion   string    `json:"descripcion" dynamodbav:"descripcion"`
	ValorAnterior string    `json:"valor_anterior" dynamodbav:"valor_anterior"`
	ValorNuevo    string    `json:"valor_nuevo" dynamodbav:"valor_nuevo"`
	UsuarioID     string    `json:"usuario_id" dynamodbav:"usuario_id"`
	FechaCambio   time.Time `json:"fecha_cambio" dynamodbav:"fecha_cambio"`
	IPAddress     string    `json:"ip_address" dynamodbav:"ip_address"`
}

// NewProveedor crea una nueva instancia de Proveedor
func NewProveedor(nombreLegal, razonSocial, identificacionFiscal string) *Proveedor {
	now := time.Now()
	return &Proveedor{
		ProveedorID:           uuid.New().String(),
		NombreLegal:           nombreLegal,
		RazonSocial:           razonSocial,
		IdentificacionFiscal:  identificacionFiscal,
		EstadoProveedor:       EstadoActivo,
		FechaRegistro:         now,
		FechaUltimaEvaluacion: now,
		Contactos:             []ContactoProveedor{},
		ProductosOfrecidos:    []ProductoOfrecido{},
		Certificaciones:       []Certificacion{},
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// NewContactoProveedor crea un nuevo contacto
func NewContactoProveedor(nombre, email, telefono, cargo string, esPrincipal bool) ContactoProveedor {
	return ContactoProveedor{
		ContactoID:          uuid.New().String(),
		Nombre:              nombre,
		Email:               email,
		Telefono:            telefono,
		Cargo:               cargo,
		EsContactoPrincipal: esPrincipal,
	}
}

// NewCertificacion crea una nueva certificación
func NewCertificacion(tipo, numero, autoridad string, fechaEmision, fechaVencimiento time.Time) Certificacion {
	estado := EstadoCertificacionActiva
	if fechaVencimiento.Before(time.Now()) {
		estado = EstadoCertificacionVencida
	} else if fechaVencimiento.Before(time.Now().AddDate(0, 0, 30)) {
		estado = EstadoCertificacionPorVencer
	}

	return Certificacion{
		TipoCertificacion: tipo,
		NumeroCertificado: numero,
		FechaEmision:      fechaEmision,
		FechaVencimiento:  fechaVencimiento,
		AutoridadEmisora:  autoridad,
		Estado:            estado,
	}
}
