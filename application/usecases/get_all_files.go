package usecases

import "github.com/yhartanto178dev/api-archiven-v2/domain"

type GetAllFilesQuery struct {
	Page    int
	PerPage int
}

type PaginatedFiles struct {
	Files      []*domain.File
	Total      int64
	Page       int
	PerPage    int
	TotalPages int64
}

type GetAllFilesUseCase struct {
	repo domain.FileRepository
}

func NewGetAllFilesUseCase(repo domain.FileRepository) *GetAllFilesUseCase {
	return &GetAllFilesUseCase{repo: repo}
}

func (uc *GetAllFilesUseCase) Execute(query GetAllFilesQuery) (*PaginatedFiles, error) {
	// Validate pagination parameters
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 || query.PerPage > 100 {
		query.PerPage = 10
	}

	skip := int64((query.Page - 1) * query.PerPage)
	limit := int64(query.PerPage)

	files, err := uc.repo.FindAll(skip, limit)
	if err != nil {
		return nil, err
	}

	total, err := uc.repo.Count()
	if err != nil {
		return nil, err
	}

	totalPages := total / int64(query.PerPage)
	if total%int64(query.PerPage) != 0 {
		totalPages++
	}

	return &PaginatedFiles{
		Files:      files,
		Total:      total,
		Page:       query.Page,
		PerPage:    query.PerPage,
		TotalPages: totalPages,
	}, nil
}
