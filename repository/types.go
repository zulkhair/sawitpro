// This file contains types that are used in the repository layer.
package repository

type GetTestByIdInput struct {
	Id string
}

type GetTestByIdOutput struct {
	Name string
}

type RegistrationInput struct {
	ID       string
	Phone    string
	Name     string
	Password string
	Salt     string
}

type RegistrationOutput struct {
	ID string
}

type User struct {
	ID       string
	Phone    string
	Name     string
	Password string
	Salt     string
}

type UpdateUser struct {
	ID    string
	Phone string
	Name  string
}

type Param struct {
	Logic    string
	Field    string
	Operator string
	Value    interface{}
}
