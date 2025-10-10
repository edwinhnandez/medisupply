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

// Constantes para los tipos de eventos
const (
	EventTypeProveedorCalificado    = "proveedor.calificado"
	EventTypeProveedorSuspendido    = "proveedor.suspendido"
	EventTypeProveedorActivado      = "proveedor.activado"
	EventTypeCertificacionPorVencer = "certificacion.por_vencer"
	EventTypeEvaluacionActualizada  = "evaluacion.actualizada"
	EventTypeCertificacionVencida   = "certificacion.vencida"
)

// Topics de eventos
const (
	TopicProveedorEvents = "supplier.events"
	TopicNotifications   = "notifications.events"
)
