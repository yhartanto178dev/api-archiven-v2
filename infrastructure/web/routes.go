package web

import (
	"github.com/labstack/echo/v4"
	"github.com/yhartanto178dev/api-archiven-v2/application/usecases"
	"github.com/yhartanto178dev/api-archiven-v2/infrastructure"
	handlers "github.com/yhartanto178dev/api-archiven-v2/infrastructure/web/handler"
	"github.com/yhartanto178dev/api-archiven-v2/infrastructure/web/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(e *echo.Echo, db *mongo.Database) { // Repository initialization
	fileRepo := infrastructure.NewMongoFileRepository(db)

	// Use cases initialization
	uploadUC := usecases.NewUploadFileUseCase(fileRepo)
	getFileUC := usecases.NewGetFileUseCase(fileRepo)
	getAllUC := usecases.NewGetAllFilesUseCase(fileRepo)

	// Handlers initialization
	fileHandlers := handlers.NewFileHandlers(uploadUC, getFileUC, getAllUC)

	// Register routes
	ApiV1 := e.Group("/api/v1")
	// Routes
	ApiV1.POST("/files", fileHandlers.UploadFile)
	ApiV1.GET("/files/:id", fileHandlers.GetFileByID)
	ApiV1.GET("/files", fileHandlers.GetAllFiles, middleware.Pagination)
	ApiV1.GET("/files/:id/download", fileHandlers.DownloadFile)
}
