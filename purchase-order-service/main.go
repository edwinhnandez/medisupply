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
	eventHandler := handlers.NewEventHandler(orderService, logger)

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

	// Suscribirse a eventos de stock
	logger.Info("Subscribing to stock events...")

	// Suscribirse a eventos de stock bajo
	err = eventBus.Subscribe(events.TopicStockEvents, "purchase-order-stock-bajo", eventHandler.HandleStockBajoEvent)
	if err != nil {
		logger.Errorf("Error subscribing to stock low events: %v", err)
	} else {
		logger.Info("Successfully subscribed to stock low events")
	}

	// Suscribirse a eventos de lote dañado
	err = eventBus.Subscribe(events.TopicStockEvents, "purchase-order-lote-danado", eventHandler.HandleLoteDanadoEvent)
	if err != nil {
		logger.Errorf("Error subscribing to damaged batch events: %v", err)
	} else {
		logger.Info("Successfully subscribed to damaged batch events")
	}

	// Suscribirse a eventos de pronóstico de alta demanda
	err = eventBus.Subscribe(events.TopicStockEvents, "purchase-order-demanda-alta", eventHandler.HandlePronosticoDemandaAltaEvent)
	if err != nil {
		logger.Errorf("Error subscribing to high demand forecast events: %v", err)
	} else {
		logger.Info("Successfully subscribed to high demand forecast events")
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
