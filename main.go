package main

import (
	"context"
	"time"

	"log"

	"github.com/labstack/echo/v4"
	"github.com/yhartanto178dev/api-archiven-v2/application/usecases"

	"github.com/yhartanto178dev/api-archiven-v2/infrastructure"
	"github.com/yhartanto178dev/api-archiven-v2/infrastructure/configs"
	handlers "github.com/yhartanto178dev/api-archiven-v2/infrastructure/web/handler"
	"github.com/yhartanto178dev/api-archiven-v2/infrastructure/web/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables
	configs.InitializeConfig()
	e := echo.New()
	// Initialize logger
	defer configs.SyncLogger()

	// Add request logging middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			latency := time.Since(start)
			configs.Logger.Infow("request",
				"method", c.Request().Method,
				"uri", c.Request().URL.Path,
				"status", c.Response().Status,
				"latency", latency.String(),
				"ip", c.RealIP(),
			)

			return err
		}
	})

	mongoURI := configs.GetMongoURI()
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable not set")
	}

	// MongoDB setup
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	dbName := configs.GetDatabaseName()
	if dbName == "" {
		log.Fatal("DATABASE_NAME environment variable not set")
	}

	db := client.Database(dbName)

	// Repository initialization
	fileRepo := infrastructure.NewMongoFileRepository(db)

	// Use cases initialization
	uploadUC := usecases.NewUploadFileUseCase(fileRepo)
	getFileUC := usecases.NewGetFileUseCase(fileRepo)
	getAllUC := usecases.NewGetAllFilesUseCase(fileRepo)

	// Handlers initialization
	fileHandlers := handlers.NewFileHandlers(uploadUC, getFileUC, getAllUC)

	// Routes
	e.POST("/files", fileHandlers.UploadFile)
	e.GET("/files/:id", fileHandlers.GetFileByID)
	e.GET("/files", fileHandlers.GetAllFiles, middleware.Pagination)
	e.GET("/files/:id/download", fileHandlers.DownloadFile)

	// Start server
	port := ":" + configs.GetServerPort()
	e.Logger.Fatal(e.Start(port))
}
