package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// This is just a test endpoint to get you started. Please delete this endpoint.
// (GET /hello)
func (s *Server) Hello(ctx echo.Context, params generated.HelloParams) error {

	var resp generated.HelloResponse
	resp.Message = fmt.Sprintf("Hello User %d", params.Id)
	return ctx.JSON(http.StatusOK, resp)
}

func (s *Server) PostRegistration(ctx echo.Context) error {
	req := new(generated.PostRegistrationJSONRequestBody)

	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Perform validation
	if !isValidPhoneNumber(*req.PhoneNumber) {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid phone number. Phone numbers must start with \"+62\" and be 10 to 13 characters in total"})
	}

	if len(*req.FullName) < 3 || len(*req.FullName) > 60 {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid full name. Full names must be 3 to 60 characters"})
	}

	if !isValidPassword(*req.Password) {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid password. Passwords must be 6 to 64 characters and contain at least 1 uppercase letter, 1 digit, and 1 special character"})
	}

	// Generate a random salt
	salt, err := generateRandomSalt()
	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	// Combine the password and salt, then hash the result
	hashedPassword, err := hashPassword(*req.Password, salt)
	if err != nil {
		log.Error(err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	output, err := s.Repository.Registration(ctx.Request().Context(), repository.RegistrationInput{
		ID:          uuid.NewString(),
		PhoneNumber: *req.PhoneNumber,
		FullName:    *req.FullName,
		Password:    hashedPassword,
		Salt:        salt,
	})

	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusInternalServerError, "Error when registering user")
	}
	return ctx.JSON(http.StatusOK, map[string]string{"message": "Registration successful", "id": output.ID})
}

func isValidPhoneNumber(phoneNumber string) bool {
	// Phone numbers must start with "+62" and be 10 to 13 characters in total
	re := regexp.MustCompile(`^\+62\d{9,11}$`)
	return re.MatchString(phoneNumber)
}

func isValidPassword(password string) bool {
	// Check length
	if len(password) < 6 || len(password) > 64 {
		return false
	}

	// Define regex patterns
	uppercasePattern := `[A-Z]`
	numberPattern := `[0-9]`
	specialCharPattern := `[^a-zA-Z0-9]`

	// Compile the regular expressions
	uppercaseRegex := regexp.MustCompile(uppercasePattern)
	numberRegex := regexp.MustCompile(numberPattern)
	specialCharRegex := regexp.MustCompile(specialCharPattern)

	// Check for at least one uppercase letter
	if !uppercaseRegex.MatchString(password) {
		return false
	}

	// Check for at least one number
	if !numberRegex.MatchString(password) {
		return false
	}

	// Check for at least one special character
	if !specialCharRegex.MatchString(password) {
		return false
	}

	// All conditions passed
	return true
}

func generateRandomSalt() (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(randomBytes), nil
}

func hashPassword(password, salt string) (string, error) {
	passwordWithSalt := []byte(password + salt)
	hash, err := bcrypt.GenerateFromPassword(passwordWithSalt, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
