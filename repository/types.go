// This file contains types that are used in the repository layer.
package repository

type GetTestByIdInput struct {
	Id string
}

type GetTestByIdOutput struct {
	Name string
}

type RegistrationInput struct {
	ID          string
	PhoneNumber string
	FullName    string
	Password    string
	Salt        string
}

type RegistrationOutput struct {
	ID string
}
