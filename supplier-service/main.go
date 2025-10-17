package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mediplus/supplier-service/internal/config"
	"mediplus/supplier-service/internal/database"
	"mediplus/supplier-service/internal/events"
	"mediplus/supplier-service/internal/handlers"
	"mediplus/supplier-service/internal/repository"
	"mediplus/supplier-service/internal/service"

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
	supplierRepo := repository.NewSupplierRepository(db, logger)
	auditRepo := repository.NewAuditRepository(db, logger)

	// Inicializar servicios
	supplierService := service.NewSupplierService(supplierRepo, auditRepo, eventBus, logger)

	// Inicializar handlers
	supplierHandler := handlers.NewSupplierHandler(supplierService, logger)
	eventHandler := handlers.NewEventHandler(supplierService, logger)

	// Configurar rutas
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Rutas de la API
	v1 := router.Group("/api/v1")
	{
		suppliers := v1.Group("/suppliers")
		{
			suppliers.POST("", supplierHandler.CreateSupplier)
			suppliers.GET("/:id", supplierHandler.GetSupplier)
			suppliers.PUT("/:id", supplierHandler.UpdateSupplier)
			suppliers.DELETE("/:id", supplierHandler.DeleteSupplier)
			suppliers.GET("", supplierHandler.ListSuppliers)
			suppliers.POST("/:id/evaluate", supplierHandler.EvaluateSupplier)
			suppliers.POST("/:id/suspend", supplierHandler.SuspendSupplier)
			suppliers.POST("/:id/activate", supplierHandler.ActivateSupplier)
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

	// Suscribirse a eventos de órdenes
	logger.Info("Subscribing to order events...")

	// Suscribirse a eventos de orden generada
	err = eventBus.Subscribe(events.TopicOrderEvents, "supplier-order-generated", eventHandler.HandleOrdenCompraGeneradaEvent)
	if err != nil {
		logger.Errorf("Error subscribing to order generated events: %v", err)
	} else {
		logger.Info("Successfully subscribed to order generated events")
	}

	// Suscribirse a eventos de orden confirmada
	err = eventBus.Subscribe(events.TopicOrderEvents, "supplier-order-confirmed", eventHandler.HandleOrdenCompraConfirmadaEvent)
	if err != nil {
		logger.Errorf("Error subscribing to order confirmed events: %v", err)
	} else {
		logger.Info("Successfully subscribed to order confirmed events")
	}

	// Suscribirse a eventos de orden recibida
	err = eventBus.Subscribe(events.TopicOrderEvents, "supplier-order-received", eventHandler.HandleOrdenCompraRecibidaEvent)
	if err != nil {
		logger.Errorf("Error subscribing to order received events: %v", err)
	} else {
		logger.Info("Successfully subscribed to order received events")
	}

	// Iniciar servidor en goroutine
	go func() {
		logger.Infof("Starting supplier service on port %s", cfg.Port)
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
