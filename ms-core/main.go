package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ms-core/internal/bootstrap"
	"ms-core/internal/config"
	"ms-core/internal/core"
	"ms-core/internal/database"
	"ms-core/internal/handlers"
	"ms-core/internal/queue"
	"ms-core/internal/routes"
	"ms-core/internal/worker"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Inicializar configuración
	cfg := config.LoadConfig()

	// Conectar a MongoDB
	database.ConnectDB(cfg.MongoURI)
	defer database.DisconnectDB()

	// Conectar a Redis
	database.ConnectRedis()

	// ---------------------------------------------------------------------------
	// Cola de tareas Redis (asynq) — broker + result store
	// ---------------------------------------------------------------------------
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	var taskBroker queue.Broker
	var resultStore queue.ResultStore

	broker, brokerErr := queue.NewAsynqBroker(redisAddr)
	if brokerErr != nil {
		log.Printf("Warning: could not create task broker: %v — async endpoints will fall back to sync", brokerErr)
	} else {
		taskBroker = broker
		defer broker.Close()
		log.Println("Task broker connected to Redis successfully")
	}

	if redisClient := database.GetRedisClient(); redisClient != nil {
		resultStore = queue.NewRedisResultStore(redisClient, queue.ResultTTL)
	} else {
		log.Println("Warning: Redis unavailable — async job results will not be persisted")
	}

	// ---------------------------------------------------------------------------
	// Context raíz para el ciclo de vida del servidor.
	// Se cancela cuando llega SIGINT o SIGTERM, lo que detiene el watcher.
	// ---------------------------------------------------------------------------
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ---------------------------------------------------------------------------
	// Sistema de plugins dinámico
	// ---------------------------------------------------------------------------
	pluginManager := core.NewPluginManager()

	if err := bootstrap.LoadPlugins(pluginManager); err != nil {
		log.Printf("Warning: error during initial plugin load: %v", err)
	}

	// ---------------------------------------------------------------------------
	// Motor de formación de equipos
	// ---------------------------------------------------------------------------
	teamFormationDeps, err := bootstrap.InitializeTeamFormation(cfg, pluginManager)
	if err != nil {
		log.Printf("Warning: Failed to initialize team formation engine: %v", err)
	}

	// ---------------------------------------------------------------------------
	// Worker de cola de tareas (goroutine independiente)
	// ---------------------------------------------------------------------------
	if taskBroker != nil && teamFormationDeps != nil {
		asynqSrv, mux := queue.NewAsynqServer(redisAddr, 16)

		tfHandler := handlers.NewTeamFormationHandler(
			teamFormationDeps.Engine,
			teamFormationDeps.Registry,
			teamFormationDeps.Plugin,
		).WithDB(database.DB).WithPluginManager(pluginManager).
			WithBroker(taskBroker).
			WithResultStore(resultStore)

		worker.NewTeamFormationWorker(tfHandler, resultStore, mux)

		go func() {
			log.Println("Starting asynq worker server...")
			if err := asynqSrv.Start(mux); err != nil {
				log.Printf("Asynq worker stopped: %v", err)
			}
		}()

		go func() {
			<-ctx.Done()
			log.Println("Stopping asynq worker server...")
			asynqSrv.Shutdown()
		}()
	}

	// ---------------------------------------------------------------------------
	// Aplicación Fiber
	// ---------------------------------------------------------------------------
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
	app.Use(logger.New())

	// Configurar rutas
	routes.SetupRoutes(app, pluginManager)

	// Configurar rutas de team formation si está inicializado
	if teamFormationDeps != nil {
		teamFormationHandler := handlers.NewTeamFormationHandler(
			teamFormationDeps.Engine,
			teamFormationDeps.Registry,
			teamFormationDeps.Plugin,
		).WithDB(database.DB).WithPluginManager(pluginManager).
			WithBroker(taskBroker).
			WithResultStore(resultStore)
		routes.SetupTeamFormationRoutes(app, teamFormationHandler)
		log.Println("Team Formation routes configured successfully")
	}

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "ms-core",
			"plugins": pluginManager.GetLoadedPlugins(),
		})
	})

	// ---------------------------------------------------------------------------
	// Arranque del servidor con graceful shutdown
	// ---------------------------------------------------------------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Port
	}

	// Goroutine que espera la señal de apagado y detiene Fiber de forma ordenada.
	go func() {
		<-ctx.Done()
		log.Println("Shutdown signal received, stopping server...")

		if err := app.Shutdown(); err != nil {
			log.Printf("Error during Fiber shutdown: %v", err)
		}

		if err := pluginManager.Shutdown(); err != nil {
			log.Printf("Error during plugin manager shutdown: %v", err)
		}
	}()

	log.Printf("ms-core service starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Printf("Fiber stopped: %v", err)
	}
}
