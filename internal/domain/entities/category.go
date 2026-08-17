package entities

import "financial-record/internal/utils"

// Category represents a financial category
type Category struct {
	ID          string
	Name        string
	Description string
	Color       string
}

// NewCategory creates a new category
func NewCategory(name string, description string, color string) *Category {
	return &Category{
		ID:          utils.GenerateID(),
		Name:        name,
		Description: description,
		Color:       color,
	}
}

// CategoryKeyword represents a keyword mapping to a category
type CategoryKeyword struct {
	ID         string
	Keyword    string
	CategoryID string
	Priority   int // Higher priority means this keyword is checked first
}

// NewCategoryKeyword creates a new category keyword
func NewCategoryKeyword(keyword string, categoryID string, priority int) *CategoryKeyword {
	return &CategoryKeyword{
		ID:         utils.GenerateID(),
		Keyword:    keyword,
		CategoryID: categoryID,
		Priority:   priority,
	}
}
