package usecases

import (
	"io"

	"github.com/yhartanto178dev/api-archiven-v2/domain"
)

type GetFileUseCase struct {
	repo domain.FileRepository
}

func NewGetFileUseCase(repo domain.FileRepository) *GetFileUseCase {
	return &GetFileUseCase{repo: repo}
}

func (uc *GetFileUseCase) Execute(id string) (*domain.File, io.ReadCloser, error) {
	return uc.repo.FindByID(id)
}
