// This file contains the interfaces for the repository layer.
// The repository layer is responsible for interacting with the database.
// For testing purpose we will generate mock implementations of these
// interfaces using mockgen. See the Makefile for more information.
package repository

import "context"

type RepositoryInterface interface {
	GetTestById(ctx context.Context, input GetTestByIdInput) (output GetTestByIdOutput, err error)
	Registration(ctx context.Context, input RegistrationInput) (output RegistrationOutput, err error)
	FindUser(ctx context.Context, params ...Param) (user User, err error)
	IncreaseLoginAttempt(ctx context.Context, phone string) (err error)
	UpdateUser(ctx context.Context, user UpdateUser) (err error)
}
