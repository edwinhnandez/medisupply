package events

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// EventBus define la interfaz para el bus de eventos
type EventBus interface {
	Publish(topic string, event interface{}) error
	Subscribe(topic string, handler func([]byte)) error
	Close() error
}

// NATSEventBus implementa EventBus usando NATS
type NATSEventBus struct {
	conn *nats.Conn
	log  *logrus.Logger
}

// NewNATSEventBus crea una nueva instancia de NATSEventBus
func NewNATSEventBus(url string) (*NATSEventBus, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &NATSEventBus{
		conn: conn,
		log:  logrus.New(),
	}, nil
}

// Publish publica un evento en el topic especificado
func (e *NATSEventBus) Publish(topic string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return e.conn.Publish(topic, data)
}

// Subscribe se suscribe a un topic y ejecuta el handler cuando llega un mensaje
func (e *NATSEventBus) Subscribe(topic string, handler func([]byte)) error {
	_, err := e.conn.Subscribe(topic, func(m *nats.Msg) {
		handler(m.Data)
	})
	return err
}

// Close cierra la conexión NATS
func (e *NATSEventBus) Close() error {
	if e.conn != nil {
		e.conn.Close()
	}
	return nil
}

// Eventos del dominio

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

// OrdenCompraEnviadaEvent se emite cuando se envía una orden de compra
type OrdenCompraEnviadaEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	OrdenID   string    `json:"orden_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		NumeroOrden string    `json:"numero_orden"`
		ProveedorID string    `json:"proveedor_id"`
		FechaEnvio  time.Time `json:"fecha_envio"`
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

// StockBajoEvent se emite cuando el stock está bajo el punto de reorden
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

// LoteDanadoEvent se emite cuando un lote es dañado por temperatura
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

// Constantes para los tipos de eventos
const (
	EventTypeOrdenCompraGenerada   = "orden_compra.generada"
	EventTypeOrdenCompraEnviada    = "orden_compra.enviada"
	EventTypeOrdenCompraConfirmada = "orden_compra.confirmada"
	EventTypeOrdenCompraRecibida   = "orden_compra.recibida"
	EventTypeStockBajo             = "stock.bajo"
	EventTypeLoteDanado            = "lote.danado"
	EventTypePronosticoDemandaAlta = "pronostico.demanda_alta"
)

// Topics de eventos
const (
	TopicOrderEvents    = "order.events"
	TopicStockEvents    = "stock.events"
	TopicSupplierEvents = "supplier.events"
)
