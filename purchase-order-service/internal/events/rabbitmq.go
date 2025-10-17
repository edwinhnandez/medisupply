package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// RabbitMQEventBus implementa EventBus usando RabbitMQ
type RabbitMQEventBus struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	log     *logrus.Logger
}

// NewRabbitMQEventBus crea una nueva instancia de RabbitMQEventBus
func NewRabbitMQEventBus(rabbitmqURL string, log *logrus.Logger) (*RabbitMQEventBus, error) {
	// Conectar a RabbitMQ
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Crear canal
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declarar exchanges
	exchanges := []string{
		TopicProveedorEvents,
		TopicNotifications,
		TopicStockEvents,
		TopicOrderEvents,
	}

	for _, exchange := range exchanges {
		err = ch.ExchangeDeclare(
			exchange, // name
			"topic",  // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}

	// Declarar colas principales
	queues := map[string]string{
		"proveedor.events":            TopicProveedorEvents,
		"notifications":               TopicNotifications,
		"stock.events":                TopicStockEvents,
		"order.events":                TopicOrderEvents,
		"proveedor.audit":             TopicProveedorEvents,
		"proveedor.evaluation":        TopicProveedorEvents,
		"purchase-order-stock-bajo":   TopicStockEvents,
		"purchase-order-lote-danado":  TopicStockEvents,
		"purchase-order-demanda-alta": TopicStockEvents,
	}

	for queueName, exchange := range queues {
		_, err = ch.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}

		// Bind queue to exchange
		err = ch.QueueBind(
			queueName, // queue name
			"#",       // routing key (wildcard)
			exchange,  // exchange
			false,
			nil,
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to bind queue %s: %w", queueName, err)
		}
	}

	log.Info("RabbitMQ EventBus initialized successfully")

	return &RabbitMQEventBus{
		conn:    conn,
		channel: ch,
		log:     log,
	}, nil
}

// Publish publica un evento en RabbitMQ
func (r *RabbitMQEventBus) Publish(topic string, event interface{}) error {
	// Serializar evento
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Determinar routing key basado en el tipo de evento
	routingKey := r.getRoutingKey(event)

	// Publicar mensaje
	err = r.channel.Publish(
		topic,      // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent, // Hacer el mensaje persistente
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	r.log.WithFields(logrus.Fields{
		"topic":       topic,
		"routing_key": routingKey,
		"event_type":  r.getEventType(event),
	}).Info("Event published successfully")

	return nil
}

// Subscribe se suscribe a eventos de un tópico específico
func (r *RabbitMQEventBus) Subscribe(topic, queueName string, handler EventHandler) error {
	// Declarar cola si no existe
	_, err := r.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = r.channel.QueueBind(
		queueName, // queue name
		"#",       // routing key
		topic,     // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Configurar QoS
	err = r.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Consumir mensajes
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Procesar mensajes en una goroutine
	go func() {
		for d := range msgs {
			r.log.WithFields(logrus.Fields{
				"queue":       queueName,
				"routing_key": d.RoutingKey,
			}).Info("Received message")

			// Procesar mensaje
			if err := handler(d.Body); err != nil {
				r.log.WithError(err).Error("Failed to process message")
				// Rechazar mensaje y no reintentar
				d.Nack(false, false)
			} else {
				// Confirmar mensaje
				d.Ack(false)
			}
		}
	}()

	r.log.WithFields(logrus.Fields{
		"topic": topic,
		"queue": queueName,
	}).Info("Subscribed to events")

	return nil
}

// Close cierra la conexión a RabbitMQ
func (r *RabbitMQEventBus) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// getRoutingKey determina el routing key basado en el tipo de evento
func (r *RabbitMQEventBus) getRoutingKey(event interface{}) string {
	switch event.(type) {
	case *ProveedorCalificadoEvent:
		return "proveedor.calificado"
	case *EvaluacionActualizadaEvent:
		return "proveedor.evaluacion.actualizada"
	case *ProveedorSuspendidoEvent:
		return "proveedor.suspendido"
	case *CertificacionPorVencerEvent:
		return "proveedor.certificacion.por_vencer"
	case *OrdenCompraGeneradaEvent:
		return "orden.generada"
	case *OrdenCompraConfirmadaEvent:
		return "orden.confirmada"
	case *OrdenCompraRecibidaEvent:
		return "orden.recibida"
	case *StockBajoEvent:
		return "stock.bajo"
	case *LoteDanadoEvent:
		return "stock.lote_danado"
	case *PronosticoDemandaAltaEvent:
		return "stock.demanda_alta"
	case *StockBajoExternoEvent:
		return "external.stock.bajo"
	case *DemandaAltaExternaEvent:
		return "external.demanda.alta"
	case *LoteDanadoExternoEvent:
		return "external.lote.danado"
	case *AlertaInventarioExternaEvent:
		return "external.alerta.inventario"
	default:
		return "unknown"
	}
}

// getEventType obtiene el tipo de evento
func (r *RabbitMQEventBus) getEventType(event interface{}) string {
	switch event.(type) {
	case *ProveedorCalificadoEvent:
		return "ProveedorCalificado"
	case *EvaluacionActualizadaEvent:
		return "EvaluacionActualizada"
	case *ProveedorSuspendidoEvent:
		return "ProveedorSuspendido"
	case *CertificacionPorVencerEvent:
		return "CertificacionPorVencer"
	case *OrdenCompraGeneradaEvent:
		return "OrdenCompraGenerada"
	case *OrdenCompraConfirmadaEvent:
		return "OrdenCompraConfirmada"
	case *OrdenCompraRecibidaEvent:
		return "OrdenCompraRecibida"
	case *StockBajoEvent:
		return "StockBajo"
	case *LoteDanadoEvent:
		return "LoteDanado"
	case *PronosticoDemandaAltaEvent:
		return "PronosticoDemandaAlta"
	case *StockBajoExternoEvent:
		return "StockBajoExterno"
	case *DemandaAltaExternaEvent:
		return "DemandaAltaExterna"
	case *LoteDanadoExternoEvent:
		return "LoteDanadoExterno"
	case *AlertaInventarioExternaEvent:
		return "AlertaInventarioExterna"
	default:
		return "Unknown"
	}
}
