package models

import (
	"time"

	"github.com/google/uuid"
)

// EstadoOrden representa el estado de una orden de compra
type EstadoOrden string

const (
	EstadoGenerada   EstadoOrden = "GENERADA"
	EstadoEnviada    EstadoOrden = "ENVIADA"
	EstadoConfirmada EstadoOrden = "CONFIRMADA"
	EstadoRecibida   EstadoOrden = "RECIBIDA"
	EstadoCancelada  EstadoOrden = "CANCELADA"
)

// Prioridad representa la prioridad de una orden
type Prioridad string

const (
	PrioridadBaja    Prioridad = "BAJA"
	PrioridadMedia   Prioridad = "MEDIA"
	PrioridadAlta    Prioridad = "ALTA"
	PrioridadCritica Prioridad = "CRITICA"
)

// EstadoItem representa el estado de un item de orden
type EstadoItem string

const (
	EstadoItemPendiente  EstadoItem = "PENDIENTE"
	EstadoItemConfirmado EstadoItem = "CONFIRMADO"
	EstadoItemRecibido   EstadoItem = "RECIBIDO"
	EstadoItemCancelado  EstadoItem = "CANCELADO"
)

// OrdenCompra representa la entidad raíz del agregado OrdenCompraAutomatica
type OrdenCompra struct {
	OrdenID          string            `json:"orden_id" dynamodbav:"orden_id"`
	NumeroOrden      string            `json:"numero_orden" dynamodbav:"numero_orden"`
	ProveedorID      string            `json:"proveedor_id" dynamodbav:"proveedor_id"`
	FechaGeneracion  time.Time         `json:"fecha_generacion" dynamodbav:"fecha_generacion"`
	EstadoOrden      EstadoOrden       `json:"estado_orden" dynamodbav:"estado_orden"`
	Prioridad        Prioridad         `json:"prioridad" dynamodbav:"prioridad"`
	MotivoGeneracion string            `json:"motivo_generacion" dynamodbav:"motivo_generacion"`
	Items            []ItemOrdenCompra `json:"items" dynamodbav:"items"`
	Evaluacion       *Evaluacion       `json:"evaluacion" dynamodbav:"evaluacion"`
	CreatedAt        time.Time         `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at" dynamodbav:"updated_at"`
}

// ItemOrdenCompra representa un item de la orden de compra
type ItemOrdenCompra struct {
	ItemID               string     `json:"item_id" dynamodbav:"item_id"`
	ProductoID           string     `json:"producto_id" dynamodbav:"producto_id"`
	CantidadSolicitada   int        `json:"cantidad_solicitada" dynamodbav:"cantidad_solicitada"`
	PrecioUnitario       float64    `json:"precio_unitario" dynamodbav:"precio_unitario"`
	TemperaturaRequerida float64    `json:"temperatura_requerida" dynamodbav:"temperatura_requerida"`
	EstadoItem           EstadoItem `json:"estado_item" dynamodbav:"estado_item"`
}

// Evaluacion representa la evaluación del proveedor para la orden
type Evaluacion struct {
	ProveedorID           string    `json:"proveedor_id" dynamodbav:"proveedor_id"`
	ScoreCalidad          float64   `json:"score_calidad" dynamodbav:"score_calidad"`
	TiempoEntregaPromedio float64   `json:"tiempo_entrega_promedio" dynamodbav:"tiempo_entrega_promedio"`
	CapacidadCadenaFrio   bool      `json:"capacidad_cadena_frio" dynamodbav:"capacidad_cadena_frio"`
	Certificaciones       []string  `json:"certificaciones" dynamodbav:"certificaciones"`
	FechaEvaluacion       time.Time `json:"fecha_evaluacion" dynamodbav:"fecha_evaluacion"`
}

// Producto representa un producto en el catálogo
type Producto struct {
	ProductoID   string       `json:"producto_id" dynamodbav:"producto_id"`
	Nombre       string       `json:"nombre" dynamodbav:"nombre"`
	StockActual  int          `json:"stock_actual" dynamodbav:"stock_actual"`
	PuntoReorden int          `json:"punto_reorden" dynamodbav:"punto_reorden"`
	StockMaximo  int          `json:"stock_maximo" dynamodbav:"stock_maximo"`
	Condiciones  *Condiciones `json:"condiciones" dynamodbav:"condiciones"`
	CreatedAt    time.Time    `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" dynamodbav:"updated_at"`
}

// Condiciones representa las condiciones requeridas para un producto
type Condiciones struct {
	CadenaFrioRequerida   bool     `json:"cadena_frio_requerida" dynamodbav:"cadena_frio_requerida"`
	TemperaturaMinima     float64  `json:"temperatura_minima" dynamodbav:"temperatura_minima"`
	TemperaturaMaxima     float64  `json:"temperatura_maxima" dynamodbav:"temperatura_maxima"`
	TiempoMaximoEntrega   int      `json:"tiempo_maximo_entrega" dynamodbav:"tiempo_maximo_entrega"`
	CondicionesRequeridas []string `json:"condiciones_requeridas" dynamodbav:"condiciones_requeridas"`
}

// ProveedorCalificado representa información de un proveedor calificado
type ProveedorCalificado struct {
	ProveedorID         string   `json:"proveedor_id" dynamodbav:"proveedor_id"`
	Certificaciones     []string `json:"certificaciones" dynamodbav:"certificaciones"`
	CapacidadCadenaFrio bool     `json:"capacidad_cadena_frio" dynamodbav:"capacidad_cadena_frio"`
	TiempoEntrega       int      `json:"tiempo_entrega" dynamodbav:"tiempo_entrega"`
	ScoreGeneral        float64  `json:"score_general" dynamodbav:"score_general"`
}

// NewOrdenCompra crea una nueva instancia de OrdenCompra
func NewOrdenCompra(proveedorID, motivoGeneracion string, prioridad Prioridad) *OrdenCompra {
	now := time.Now()
	return &OrdenCompra{
		OrdenID:          uuid.New().String(),
		NumeroOrden:      generateOrderNumber(),
		ProveedorID:      proveedorID,
		FechaGeneracion:  now,
		EstadoOrden:      EstadoGenerada,
		Prioridad:        prioridad,
		MotivoGeneracion: motivoGeneracion,
		Items:            []ItemOrdenCompra{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// NewItemOrdenCompra crea un nuevo item de orden
func NewItemOrdenCompra(productoID string, cantidad int, precioUnitario, temperaturaRequerida float64) ItemOrdenCompra {
	return ItemOrdenCompra{
		ItemID:               uuid.New().String(),
		ProductoID:           productoID,
		CantidadSolicitada:   cantidad,
		PrecioUnitario:       precioUnitario,
		TemperaturaRequerida: temperaturaRequerida,
		EstadoItem:           EstadoItemPendiente,
	}
}

// NewProducto crea una nueva instancia de Producto
func NewProducto(nombre string, stockActual, puntoReorden, stockMaximo int) *Producto {
	now := time.Now()
	return &Producto{
		ProductoID:   uuid.New().String(),
		Nombre:       nombre,
		StockActual:  stockActual,
		PuntoReorden: puntoReorden,
		StockMaximo:  stockMaximo,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewCondiciones crea nuevas condiciones para un producto
func NewCondiciones(cadenaFrioRequerida bool, tempMin, tempMax float64, tiempoMaxEntrega int, condicionesRequeridas []string) *Condiciones {
	return &Condiciones{
		CadenaFrioRequerida:   cadenaFrioRequerida,
		TemperaturaMinima:     tempMin,
		TemperaturaMaxima:     tempMax,
		TiempoMaximoEntrega:   tiempoMaxEntrega,
		CondicionesRequeridas: condicionesRequeridas,
	}
}

// generateOrderNumber genera un número de orden único
func generateOrderNumber() string {
	return "ORD-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:8]
}

// AddItem agrega un item a la orden
func (o *OrdenCompra) AddItem(item ItemOrdenCompra) {
	o.Items = append(o.Items, item)
	o.UpdatedAt = time.Now()
}

// ConfirmOrder confirma la orden
func (o *OrdenCompra) ConfirmOrder() {
	o.EstadoOrden = EstadoConfirmada
	o.UpdatedAt = time.Now()
}

// ReceiveOrder marca la orden como recibida
func (o *OrdenCompra) ReceiveOrder() {
	o.EstadoOrden = EstadoRecibida
	o.UpdatedAt = time.Now()
}

// SendOrder marca la orden como enviada
func (o *OrdenCompra) SendOrder() {
	o.EstadoOrden = EstadoEnviada
	o.UpdatedAt = time.Now()
}

// CancelOrder cancela la orden
func (o *OrdenCompra) CancelOrder() {
	o.EstadoOrden = EstadoCancelada
	o.UpdatedAt = time.Now()
}

// IsLowStock verifica si el producto está bajo stock
func (p *Producto) IsLowStock() bool {
	return p.StockActual <= p.PuntoReorden
}

// NeedsReorder verifica si el producto necesita reorden
func (p *Producto) NeedsReorder() bool {
	return p.StockActual < p.PuntoReorden
}

// CanReorder verifica si se puede hacer reorden (no exceder stock máximo)
func (p *Producto) CanReorder(cantidad int) bool {
	return p.StockActual+cantidad <= p.StockMaximo
}
