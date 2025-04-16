package handlers

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yhartanto178dev/api-archiven-v2/application/usecases"
	"github.com/yhartanto178dev/api-archiven-v2/domain"
	"github.com/yhartanto178dev/api-archiven-v2/infrastructure/configs"
	"github.com/yhartanto178dev/api-archiven-v2/infrastructure/web/responses"
)

type FileHandlers struct {
	uploadUseCase  *usecases.UploadFileUseCase
	getFileUseCase *usecases.GetFileUseCase
	getAllUseCase  *usecases.GetAllFilesUseCase
}

func NewFileHandlers(
	uploadUC *usecases.UploadFileUseCase,
	getUC *usecases.GetFileUseCase,
	getAllUC *usecases.GetAllFilesUseCase,
) *FileHandlers {
	return &FileHandlers{
		uploadUseCase:  uploadUC,
		getFileUseCase: getUC,
		getAllUseCase:  getAllUC,
	}
}

func (h *FileHandlers) UploadFile(c echo.Context) error {
	// Validate file existence
	// req := domain.FileUploadRequest{}
	// if err := c.Bind(&req); err != nil {
	// 	return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	// }

	// if err := domain.ValidateUploadRequest(req); err != nil {
	// 	return c.JSON(http.StatusBadRequest, map[string]interface{}{
	// 		"error":   "validation failed",
	// 		"details": domain.FormatValidationErrors(err),
	// 	})
	// }
	// Get the file from the form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file is required"})
	}

	// Validate file size
	if err := domain.ValidateFileSize(fileHeader.Size); err != nil {
		configs.Logger.Errorw("file upload failed",
			"error", err.Error(),
			"filename", fileHeader.Filename,
			"content_type", fileHeader.Header.Get("Content-Type"),
			"time", time.Now().Format(time.RFC3339),
		)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "file validation failed",
			"details": domain.FormatValidationErrors(err),
		})
	}

	// Validate content type
	if err := domain.ValidateMimeType(fileHeader.Header.Get("Content-Type")); err != nil {
		configs.Logger.Errorw("file upload failed",
			"error", err.Error(),
			"filename", fileHeader.Filename,
			"content_type", fileHeader.Header.Get("Content-Type"),
			"time", time.Now().Format(time.RFC3339),
		)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "file type not allowed",
			"details": domain.FormatValidationErrors(err),
		})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to open file"})
	}
	defer src.Close()
	//	// Create a new file entity
	cmd := usecases.UploadFileCommand{
		Name:        fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Content:     src,
	}
	// Save the file using the use case
	uploadedFile, err := h.uploadUseCase.Execute(cmd)
	if err != nil {
		configs.Logger.Errorw("file upload failed",
			"error", err.Error(),
			"filename", fileHeader.Filename,
			"content_type", fileHeader.Header.Get("Content-Type"),
			"time", time.Now().Format(time.RFC3339),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	/// Logger
	configs.Logger.Infow("file uploaded successfully",
		"file_id", uploadedFile.ID,
		"file_name", uploadedFile.Name,
		"file_size", uploadedFile.Size,
	)
	fileResponse := responses.FileResponse{
		Name:        uploadedFile.Name,
		Size:        uploadedFile.Size,
		ContentType: uploadedFile.ContentType,
		UploadDate:  uploadedFile.UploadDate.Format(time.RFC3339),
		DownloadURL: "http://localhost:8080/files/" + uploadedFile.ID + "/download",
	}

	return c.JSON(http.StatusCreated, fileResponse)
}

func (h *FileHandlers) GetFileByID(c echo.Context) error {
	id := c.Param("id")
	file, _, err := h.getFileUseCase.Execute(id)
	if err != nil {
		if err == domain.ErrFileNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	fileResponse := responses.FileResponse{
		Name:        file.Name,
		Size:        file.Size,
		ContentType: file.ContentType,
		UploadDate:  file.UploadDate.Format(time.RFC3339),
		DownloadURL: "http://localhost:8080/files/" + file.ID + "/download",
	}

	return c.JSON(http.StatusOK, fileResponse)
}

func (h *FileHandlers) GetAllFiles(c echo.Context) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	result, err := h.getAllUseCase.Execute(usecases.GetAllFilesQuery{
		Page:    page,
		PerPage: perPage,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	response := map[string]interface{}{
		"data": responses.BuildFilesResponse(result.Files, c),
		"pagination": map[string]interface{}{
			"total":       result.Total,
			"page":        result.Page,
			"per_page":    result.PerPage,
			"total_pages": result.TotalPages,
		},
	}

	return c.JSON(http.StatusOK, response)
}

func (h *FileHandlers) DownloadFile(c echo.Context) error {
	id := c.Param("id")
	file, content, err := h.getFileUseCase.Execute(id)
	if err != nil {
		if err == domain.ErrFileNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer content.Close()

	c.Response().Header().Set("Content-Type", file.ContentType)
	c.Response().Header().Set("Content-Disposition", "attachment; filename=\""+file.Name+"\"")
	c.Response().Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))

	_, err = io.Copy(c.Response().Writer, content)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to stream file"})
	}

	return nil
}
