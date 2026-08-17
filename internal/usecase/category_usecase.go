package usecase

import (
	"financial-record/internal/domain/entities"
	"financial-record/internal/domain/repository"
)

// CategoryUseCase handles business logic for categories
type CategoryUseCase struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryUseCase creates a new category use case
func NewCategoryUseCase(categoryRepo repository.CategoryRepository) *CategoryUseCase {
	return &CategoryUseCase{
		categoryRepo: categoryRepo,
	}
}

// CategorizeTransaction categorizes a transaction based on its description
func (uc *CategoryUseCase) CategorizeTransaction(description string) (string, error) {
	return uc.categoryRepo.CategorizeTransaction(description)
}

// GetAllCategories retrieves all categories
func (uc *CategoryUseCase) GetAllCategories() ([]*entities.Category, error) {
	return uc.categoryRepo.GetAllCategories()
}

// CreateCategory creates a new category
func (uc *CategoryUseCase) CreateCategory(name, description, color string) error {
	category := entities.NewCategory(name, description, color)
	return uc.categoryRepo.CreateCategory(category)
}

// GetAllKeywords retrieves all category keywords
func (uc *CategoryUseCase) GetAllKeywords() ([]*entities.CategoryKeyword, error) {
	return uc.categoryRepo.GetAllKeywords()
}

// CreateKeyword creates a new category keyword
func (uc *CategoryUseCase) CreateKeyword(keyword, categoryID string, priority int) error {
	keywordEntity := entities.NewCategoryKeyword(keyword, categoryID, priority)
	return uc.categoryRepo.CreateKeyword(keywordEntity)
}
