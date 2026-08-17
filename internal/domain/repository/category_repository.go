package repository

import (
	"financial-record/internal/domain/entities"
)

// CategoryRepository defines the interface for category data operations
type CategoryRepository interface {
	// GetAllCategories retrieves all categories
	GetAllCategories() ([]*entities.Category, error)

	// CreateCategory creates a new category
	CreateCategory(category *entities.Category) error

	// GetAllKeywords retrieves all category keywords
	GetAllKeywords() ([]*entities.CategoryKeyword, error)

	// CreateKeyword creates a new category keyword
	CreateKeyword(keyword *entities.CategoryKeyword) error

	// CategorizeTransaction categorizes a transaction based on its description
	CategorizeTransaction(description string) (string, error)

	// AddDefaultCategories adds default categories for Indonesian context
	AddDefaultCategories() error
}
