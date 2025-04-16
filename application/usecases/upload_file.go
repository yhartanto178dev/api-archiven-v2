package usecases

import (
	"io"
	"time"

	"github.com/yhartanto178dev/api-archiven-v2/domain"
)

type UploadFileCommand struct {
	Name        string
	ContentType string
	Content     io.Reader
}

type UploadFileUseCase struct {
	repo domain.FileRepository
}

func NewUploadFileUseCase(repo domain.FileRepository) *UploadFileUseCase {
	return &UploadFileUseCase{repo: repo}
}

func (uc *UploadFileUseCase) Execute(command UploadFileCommand) (*domain.File, error) {
	file := &domain.File{
		Name:        command.Name,
		ContentType: command.ContentType,
		UploadDate:  time.Now(),
	}
	err := uc.repo.Save(file, command.Content)
	return file, err
}
