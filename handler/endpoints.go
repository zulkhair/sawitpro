package handler

import (
	"errors"
	"fmt"
	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

// Todo : create standard response helper

// This is just a test endpoint to get you started. Please delete this endpoint.
// (GET /hello)
func (s *Server) Hello(ctx echo.Context, params generated.HelloParams) error {
	var resp generated.HelloResponse
	resp.Message = fmt.Sprintf("Hello User %d", params.Id)
	return ctx.JSON(http.StatusOK, resp)
}

// PostRegistration : Handler for registering new user
func (s *Server) PostRegistration(ctx echo.Context) error {
	req := new(generated.PostRegistrationJSONRequestBody)

	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Perform validation
	if !isValidPhoneNumber(req.Phone) {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid phone number. Phone numbers must start with +62 and be 10 to 13 characters in total"})
	}

	if len(req.Name) < 3 || len(req.Name) > 60 {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid full name. Full names must be 3 to 60 characters"})
	}

	if !isValidPassword(req.Password) {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid password. Passwords must be 6 to 64 characters and contain at least 1 uppercase letter, 1 digit, and 1 special character"})
	}

	// Generate a random salt
	salt, err := generateRandomSalt()
	if err != nil { // Todo : make this function as interface, this error cannot covered by unit test by now
		log.Error(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	// Combine the password and salt, then hash the result
	hashedPassword, err := hashPassword(req.Password, salt)
	if err != nil { // Todo : make this function as interface, this error cannot covered by unit test by now
		log.Error(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	output, err := s.Repository.Registration(ctx.Request().Context(), repository.RegistrationInput{
		ID:       uuid.NewString(),
		Phone:    req.Phone,
		Name:     req.Name,
		Password: hashedPassword,
		Salt:     salt,
	})

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Check if the error code is 23505 (unique violation)
			if pqErr.Code == "23505" {
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Phone number already exist"})
			}
		}
		log.Error(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Error when registering user"})
	}
	return ctx.JSON(http.StatusOK, map[string]string{"message": "Registration successful", "id": output.ID})
}

// PostLogin : This handler is for login, returning jwt token
func (s *Server) PostLogin(ctx echo.Context) error {
	req := new(generated.PostLoginJSONRequestBody)

	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Perform validation
	if !isValidPhoneNumber(req.Phone) {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid phone number. Phone numbers must start with +62 and be 10 to 13 characters in total"})
	}

	if !isValidPassword(req.Password) {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid password. Passwords must be 6 to 64 characters and contain at least 1 uppercase letter, 1 digit, and 1 special character"})
	}

	// Find user by phone to database
	user, err := s.Repository.FindUser(ctx.Request().Context(), repository.Param{
		Logic:    "AND",
		Field:    "phone",
		Operator: "=",
		Value:    req.Phone,
	})
	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User not found"})
	}

	// Compare password
	passwordWithSalt := req.Password + user.Salt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordWithSalt)); err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid password"})
	}

	// create jwt token
	exp := time.Now().Add(time.Hour * 1)
	token, err := createToken(user.ID, exp)
	if err != nil { // Todo : make this function as interface, this error cannot covered by unit test by now
		log.Error(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	// Login attempt increment
	err = s.Repository.IncreaseLoginAttempt(ctx.Request().Context(), req.Phone)
	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Login successful", "token": token, "phone": user.Phone})
}

// GetProfile : this handler is for getting profile of user
func (s *Server) GetProfile(ctx echo.Context) error {
	// Todo : create middleware to check the token
	// Validate token
	ID, err := validateToken(ctx)
	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden code"})
	}

	// Find user by ID
	user, err := s.Repository.FindUser(ctx.Request().Context(), repository.Param{
		Logic:    "AND",
		Field:    "id",
		Operator: "=",
		Value:    ID,
	})
	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User not found"})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"phone": user.Phone, "name": user.Name})
}

func (s *Server) PutProfile(ctx echo.Context) error {
	// Todo : create middleware to check the token
	// Validate token
	ID, err := validateToken(ctx)
	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden code"})
	}

	req := new(generated.PutProfileJSONRequestBody)

	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Find user by ID
	user, err := s.Repository.FindUser(ctx.Request().Context(), repository.Param{
		Logic:    "AND",
		Field:    "id",
		Operator: "=",
		Value:    ID,
	})
	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User not found"})
	}

	userUpdate := repository.UpdateUser{}
	userUpdate.ID = user.ID
	if req.Phone != nil {
		// Perform validation
		if !isValidPhoneNumber(*req.Phone) { // Todo : create function for standard response, because some response are similar
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid phone number. Phone numbers must start with +62 and be 10 to 13 characters in total"})
		}

		// set to new value
		userUpdate.Phone = *req.Phone
	} else {
		// set to old value
		userUpdate.Phone = user.Phone
	}

	if req.Name != nil {
		// Perform validation
		if len(*req.Name) < 3 || len(*req.Name) > 60 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid full name. Full names must be 3 to 60 characters"})
		}

		// set to new value
		userUpdate.Name = *req.Name
	} else {
		// set to old value
		userUpdate.Name = user.Name
	}

	// Update user
	err = s.Repository.UpdateUser(ctx.Request().Context(), userUpdate)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Check if the error code is 23505 (unique violation)
			if pqErr.Code == "23505" {
				return ctx.JSON(http.StatusConflict, map[string]string{"error": "Phone number already exist"})
			}
		}
		log.Error(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Error when registering user"})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "User updated"})
}
