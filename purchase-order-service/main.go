package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mediplus/purchase-order-service/internal/config"
	"mediplus/purchase-order-service/internal/database"
	"mediplus/purchase-order-service/internal/events"
	"mediplus/purchase-order-service/internal/handlers"
	"mediplus/purchase-order-service/internal/repository"
	"mediplus/purchase-order-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Configurar logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Cargar configuración
	cfg := config.Load()

	// Inicializar DynamoDB
	db, err := database.NewDynamoDBClient(cfg.AWSRegion, cfg.DynamoDBEndpoint)
	if err != nil {
		logger.Fatalf("Error connecting to DynamoDB: %v", err)
	}

	// Inicializar RabbitMQ para eventos
	eventBus, err := events.NewRabbitMQEventBus(cfg.RabbitMQURL, logger)
	if err != nil {
		logger.Fatalf("Error connecting to RabbitMQ: %v", err)
	}
	defer eventBus.Close()

	// Inicializar repositorios
	orderRepo := repository.NewOrderRepository(db, logger)
	productRepo := repository.NewProductRepository(db, logger)

	// Inicializar servicios
	orderService := service.NewOrderService(orderRepo, productRepo, eventBus, logger)

	// Inicializar handlers
	orderHandler := handlers.NewOrderHandler(orderService, logger)
	externalEventHandler := handlers.NewExternalEventHandler(orderService, logger)
	externalSimulatorHandler := handlers.NewExternalSimulatorHandler(eventBus, logger)

	// Suscribirse a eventos externos para creación automática de órdenes
	err = eventBus.Subscribe(events.TopicExternalEvents, "purchase-order-external-events", externalEventHandler.HandleExternalEvent)
	if err != nil {
		logger.Fatalf("Error subscribing to external events: %v", err)
	}
	logger.Info("Subscribed to external events for auto order generation")

	// Configurar rutas
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Rutas de la API
	v1 := router.Group("/api/v1")
	{
		orders := v1.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.PUT("/:id", orderHandler.UpdateOrder)
			orders.DELETE("/:id", orderHandler.DeleteOrder)
			orders.GET("", orderHandler.ListOrders)
			orders.POST("/:id/confirm", orderHandler.ConfirmOrder)
			orders.POST("/:id/receive", orderHandler.ReceiveOrder)
			orders.POST("/auto-generate", orderHandler.AutoGenerateOrder)
		}

		// Rutas para simulación de eventos externos
		external := v1.Group("/external")
		{
			external.GET("/event-types", externalSimulatorHandler.GetExternalEventTypes)
			simulate := external.Group("/simulate")
			{
				simulate.POST("/stock-bajo", externalSimulatorHandler.SimulateStockBajoExterno)
				simulate.POST("/demanda-alta", externalSimulatorHandler.SimulateDemandaAltaExterna)
				simulate.POST("/lote-danado", externalSimulatorHandler.SimulateLoteDanadoExterno)
				simulate.POST("/alerta-inventario", externalSimulatorHandler.SimulateAlertaInventarioExterna)
			}
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Configurar servidor
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Iniciar servidor en goroutine
	go func() {
		logger.Infof("Starting purchase order service on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Esperar señal de terminación
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Cerrar servidor gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
