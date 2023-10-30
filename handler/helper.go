package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
	"time"
)

var secret = []byte("33cfdeb6-a200-483d-84c8-ac0242682a40") // Todo : move this to config file or env variable or database
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

func createToken(id string, exp time.Time) (string, error) {
	// Create a new token object, specifying the signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": exp.Unix(), // Token expires in 1 hour
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Todo : this should be on middleware, but still don't know how to generate code using deepmap open api codegen with jwt auth support
func validateToken(ctx echo.Context) (string, error) {
	// Get authorization header
	authorization := ctx.Request().Header.Get("Authorization")
	if authorization == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	// Extract the token from the Authorization header
	tokenString := strings.Replace(authorization, "Bearer ", "", 1)

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return "", err
	}

	// Check token validity
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	return claims["id"].(string), nil
}
