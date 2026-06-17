package entities

// User represents a user in the system
type User struct {
	Phone  string
	Name   string
	Balance float64
}

// NewUser creates a new user
func NewUser(phone, name string, initialBalance float64) *User {
	return &User{
		Phone:   phone,
		Name:    name,
		Balance: initialBalance,
	}
}
