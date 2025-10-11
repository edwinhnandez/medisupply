package events

import (
	"time"
)

// EventHandler define el tipo de función para manejar eventos
type EventHandler func([]byte) error

// EventBus define la interfaz para el bus de eventos
type EventBus interface {
	Publish(topic string, event interface{}) error
	Subscribe(topic, queueName string, handler EventHandler) error
	Close() error
}

// Topics de eventos
const (
	TopicProveedorEvents = "supplier.events"
	TopicNotifications   = "notifications.events"
	TopicStockEvents     = "stock.events"
	TopicOrderEvents     = "order.events"
	TopicExternalEvents  = "external.events"
)

// Eventos del dominio

// ProveedorCalificadoEvent se emite cuando un proveedor es calificado
type ProveedorCalificadoEvent struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	ProveedorID string    `json:"proveedor_id"`
	Timestamp   time.Time `json:"timestamp"`
	Data        struct {
		NombreLegal         string   `json:"nombre_legal"`
		RazonSocial         string   `json:"razon_social"`
		ScoreGeneral        float64  `json:"score_general"`
		Certificaciones     []string `json:"certificaciones"`
		CapacidadCadenaFrio bool     `json:"capacidad_cadena_frio"`
	} `json:"data"`
}

// ProveedorSuspendidoEvent se emite cuando un proveedor es suspendido
type ProveedorSuspendidoEvent struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	ProveedorID string    `json:"proveedor_id"`
	Timestamp   time.Time `json:"timestamp"`
	Data        struct {
		MotivoSuspension string    `json:"motivo_suspension"`
		FechaSuspension  time.Time `json:"fecha_suspension"`
	} `json:"data"`
}

// CertificacionPorVencerEvent se emite cuando una certificación está por vencer
type CertificacionPorVencerEvent struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	ProveedorID string    `json:"proveedor_id"`
	Timestamp   time.Time `json:"timestamp"`
	Data        struct {
		CertificacionID   string    `json:"certificacion_id"`
		TipoCertificacion string    `json:"tipo_certificacion"`
		FechaVencimiento  time.Time `json:"fecha_vencimiento"`
		DiasRestantes     int       `json:"dias_restantes"`
	} `json:"data"`
}

// EvaluacionActualizadaEvent se emite cuando se actualiza la evaluación de un proveedor
type EvaluacionActualizadaEvent struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	ProveedorID string    `json:"proveedor_id"`
	Timestamp   time.Time `json:"timestamp"`
	Data        struct {
		ScoreAnterior        float64 `json:"score_anterior"`
		ScoreNuevo           float64 `json:"score_nuevo"`
		CumplimientoPlazos   float64 `json:"cumplimiento_plazos"`
		CalidadProductos     float64 `json:"calidad_productos"`
		RespuestaEmergencias float64 `json:"respuesta_emergencias"`
	} `json:"data"`
}

// OrdenCompraGeneradaEvent se emite cuando se genera una orden de compra
type OrdenCompraGeneradaEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	OrdenID   string    `json:"orden_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		NumeroOrden      string  `json:"numero_orden"`
		ProveedorID      string  `json:"proveedor_id"`
		MotivoGeneracion string  `json:"motivo_generacion"`
		Prioridad        string  `json:"prioridad"`
		TotalItems       int     `json:"total_items"`
		ValorTotal       float64 `json:"valor_total"`
	} `json:"data"`
}

// OrdenCompraConfirmadaEvent se emite cuando se confirma una orden de compra
type OrdenCompraConfirmadaEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	OrdenID   string    `json:"orden_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		NumeroOrden       string    `json:"numero_orden"`
		ProveedorID       string    `json:"proveedor_id"`
		FechaConfirmacion time.Time `json:"fecha_confirmacion"`
	} `json:"data"`
}

// OrdenCompraRecibidaEvent se emite cuando se recibe una orden de compra
type OrdenCompraRecibidaEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	OrdenID   string    `json:"orden_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		NumeroOrden    string    `json:"numero_orden"`
		ProveedorID    string    `json:"proveedor_id"`
		FechaRecepcion time.Time `json:"fecha_recepcion"`
	} `json:"data"`
}

// StockBajoEvent se emite cuando el stock de un producto está bajo
type StockBajoEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	ProductoID string    `json:"producto_id"`
	Timestamp  time.Time `json:"timestamp"`
	Data       struct {
		NombreProducto    string `json:"nombre_producto"`
		StockActual       int    `json:"stock_actual"`
		PuntoReorden      int    `json:"punto_reorden"`
		StockMaximo       int    `json:"stock_maximo"`
		CantidadRequerida int    `json:"cantidad_requerida"`
	} `json:"data"`
}

// LoteDanadoEvent se emite cuando se detecta un lote dañado
type LoteDanadoEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	ProductoID string    `json:"producto_id"`
	Timestamp  time.Time `json:"timestamp"`
	Data       struct {
		NombreProducto        string  `json:"nombre_producto"`
		LoteID                string  `json:"lote_id"`
		CantidadDanada        int     `json:"cantidad_danada"`
		TemperaturaRegistrada float64 `json:"temperatura_registrada"`
		TemperaturaRequerida  float64 `json:"temperatura_requerida"`
		MotivoDanio           string  `json:"motivo_danio"`
	} `json:"data"`
}

// PronosticoDemandaAltaEvent se emite cuando se pronostica alta demanda
type PronosticoDemandaAltaEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	ProductoID string    `json:"producto_id"`
	Timestamp  time.Time `json:"timestamp"`
	Data       struct {
		NombreProducto      string  `json:"nombre_producto"`
		DemandaPronosticada int     `json:"demanda_pronosticada"`
		StockActual         int     `json:"stock_actual"`
		CantidadRequerida   int     `json:"cantidad_requerida"`
		ConfianzaPronostico float64 `json:"confianza_pronostico"`
	} `json:"data"`
}

// Eventos externos que pueden disparar creación automática de órdenes

// StockBajoExternoEvent se emite por sistemas externos cuando el stock está bajo
type StockBajoExternoEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	ProductoID string    `json:"producto_id"`
	Timestamp  time.Time `json:"timestamp"`
	Source     string    `json:"source"` // Sistema que emite el evento
	Data       struct {
		NombreProducto    string `json:"nombre_producto"`
		StockActual       int    `json:"stock_actual"`
		PuntoReorden      int    `json:"punto_reorden"`
		StockMaximo       int    `json:"stock_maximo"`
		CantidadRequerida int    `json:"cantidad_requerida"`
		Prioridad         string `json:"prioridad"`
		Urgencia          string `json:"urgencia"`
	} `json:"data"`
}

// DemandaAltaExternaEvent se emite por sistemas de pronóstico externos
type DemandaAltaExternaEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	ProductoID string    `json:"producto_id"`
	Timestamp  time.Time `json:"timestamp"`
	Source     string    `json:"source"` // Sistema que emite el evento
	Data       struct {
		NombreProducto      string  `json:"nombre_producto"`
		DemandaPronosticada int     `json:"demanda_pronosticada"`
		StockActual         int     `json:"stock_actual"`
		CantidadRequerida   int     `json:"cantidad_requerida"`
		ConfianzaPronostico float64 `json:"confianza_pronostico"`
		PeriodoPronostico   string  `json:"periodo_pronostico"`
		Prioridad           string  `json:"prioridad"`
	} `json:"data"`
}

// LoteDanadoExternoEvent se emite por sistemas de monitoreo externos
type LoteDanadoExternoEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	ProductoID string    `json:"producto_id"`
	Timestamp  time.Time `json:"timestamp"`
	Source     string    `json:"source"` // Sistema que emite el evento
	Data       struct {
		NombreProducto        string  `json:"nombre_producto"`
		LoteID                string  `json:"lote_id"`
		CantidadDanada        int     `json:"cantidad_danada"`
		TemperaturaRegistrada float64 `json:"temperatura_registrada"`
		TemperaturaRequerida  float64 `json:"temperatura_requerida"`
		MotivoDanio           string  `json:"motivo_danio"`
		Urgencia              string  `json:"urgencia"`
	} `json:"data"`
}

// AlertaInventarioExternaEvent se emite por sistemas de inventario externos
type AlertaInventarioExternaEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	ProductoID string    `json:"producto_id"`
	Timestamp  time.Time `json:"timestamp"`
	Source     string    `json:"source"` // Sistema que emite el evento
	Data       struct {
		NombreProducto    string `json:"nombre_producto"`
		TipoAlerta        string `json:"tipo_alerta"`
		Descripcion       string `json:"descripcion"`
		StockActual       int    `json:"stock_actual"`
		StockMinimo       int    `json:"stock_minimo"`
		CantidadRequerida int    `json:"cantidad_requerida"`
		Prioridad         string `json:"prioridad"`
		FechaVencimiento  string `json:"fecha_vencimiento,omitempty"`
	} `json:"data"`
}

// Constantes para los tipos de eventos
const (
	EventTypeProveedorCalificado    = "proveedor.calificado"
	EventTypeProveedorSuspendido    = "proveedor.suspendido"
	EventTypeProveedorActivado      = "proveedor.activado"
	EventTypeCertificacionPorVencer = "certificacion.por_vencer"
	EventTypeEvaluacionActualizada  = "evaluacion.actualizada"
	EventTypeCertificacionVencida   = "certificacion.vencida"
	EventTypeOrdenCompraGenerada    = "orden.generada"
	EventTypeOrdenCompraConfirmada  = "orden.confirmada"
	EventTypeOrdenCompraRecibida    = "orden.recibida"
	EventTypeStockBajo              = "stock.bajo"
	EventTypeLoteDanado             = "stock.lote_danado"
	EventTypePronosticoDemandaAlta  = "stock.demanda_alta"
	// Eventos externos
	EventTypeStockBajoExterno        = "external.stock.bajo"
	EventTypeDemandaAltaExterna      = "external.demanda.alta"
	EventTypeLoteDanadoExterno       = "external.lote.danado"
	EventTypeAlertaInventarioExterna = "external.alerta.inventario"
)
