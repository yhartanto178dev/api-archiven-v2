package responses

import (
	"github.com/labstack/echo/v4"
	"github.com/yhartanto178dev/api-archiven-v2/domain"
)

type FileResponse struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	UploadDate  string `json:"upload_date"`
	DownloadURL string `json:"download_url"`
}

func BuildFilesResponse(files []*domain.File, c echo.Context) []FileResponse {
	response := make([]FileResponse, len(files))
	for i, file := range files {
		response[i] = FileResponse{
			Name:        file.Name,
			Size:        file.Size,
			ContentType: file.ContentType,
			UploadDate:  file.UploadDate.Format("2006-01-02"),
			DownloadURL: c.Scheme() + "://" + c.Request().Host + "/files/" + file.ID,
		}
	}
	return response
}
