package handler

import (
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHello(t *testing.T) {
	// Create a new Echo instance
	e := echo.New()

	// Create a request
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock the generated.HelloParams
	params := generated.HelloParams{
		Id: 123,
	}

	// Create a new instance of your server
	s := &Server{}

	// Call the Hello handler
	err := s.Hello(c, params)

	// Assert that there is no error
	assert.NoError(t, err)

	// Assert the HTTP status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Assert the response body
	expectedResponse := fmt.Sprintf("{\"message\":\"Hello User %d\"}\n", params.Id)
	assert.Equal(t, expectedResponse, rec.Body.String())
}

func TestPostRegistration(t *testing.T) {
	// Mock
	type fields struct {
		repo *repository.MockRepositoryInterface
	}

	// Output parameters
	type want struct {
		httpStatus int
		content    string
	}

	// Test Case
	tests := []struct {
		prepare func(f *fields)
		name    string
		args    string
		want    want
		wantErr bool
	}{
		{
			name: "Success",
			prepare: func(f *fields) {
				f.repo.EXPECT().Registration(gomock.Any(), gomock.Any()).Return(repository.RegistrationOutput{ID: "123"}, nil)
			},
			args: fmt.Sprintf(`{"phone": "%s", "name": "%s", "password": "%s"}`, "+62856712332", "Success User", "Password1!"),
			want: want{
				httpStatus: http.StatusOK,
				content:    "{\"id\":\"123\",\"message\":\"Registration successful\"}\n",
			},
			wantErr: false,
		}, {
			name: "Invalid request payload",
			prepare: func(f *fields) {

			},
			args: fmt.Sprintf("asd"),
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Invalid request payload\"}\n",
			},
			wantErr: false,
		}, {
			name: "Invalid phone number",
			prepare: func(f *fields) {

			},
			args: fmt.Sprintf(`{"phone": "%s", "name": "%s", "password": "%s"}`, "123", "User", "Password1!"),
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Invalid phone number. Phone numbers must start with +62 and be 10 to 13 characters in total\"}\n",
			},
			wantErr: false,
		}, {
			name: "Invalid full name",
			prepare: func(f *fields) {

			},
			args: fmt.Sprintf(`{"phone": "%s", "name": "%s", "password": "%s"}`, "+62856712332", "", "Password1!"),
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Invalid full name. Full names must be 3 to 60 characters\"}\n",
			},
			wantErr: false,
		}, {
			name: "Invalid password",
			prepare: func(f *fields) {

			},
			args: fmt.Sprintf(`{"phone": "%s", "name": "%s", "password": "%s"}`, "+62856712332", "User", ""),
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Invalid password. Passwords must be 6 to 64 characters and contain at least 1 uppercase letter, 1 digit, and 1 special character\"}\n",
			},
			wantErr: false,
		}, {
			name: "Phone number already exist",
			prepare: func(f *fields) {
				f.repo.EXPECT().Registration(gomock.Any(), gomock.Any()).Return(repository.RegistrationOutput{}, &pq.Error{Code: "23505"})
			},
			args: fmt.Sprintf(`{"phone": "%s", "name": "%s", "password": "%s"}`, "+62856712332", "User", "Password1!"),
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Phone number already exist\"}\n",
			},
			wantErr: false,
		}, {
			name: "Error when registering user",
			prepare: func(f *fields) {
				f.repo.EXPECT().Registration(gomock.Any(), gomock.Any()).Return(repository.RegistrationOutput{}, fmt.Errorf("error"))
			},
			args: fmt.Sprintf(`{"phone": "%s", "name": "%s", "password": "%s"}`, "+62856712332", "User", "Password1!"),
			want: want{
				httpStatus: http.StatusInternalServerError,
				content:    "{\"error\":\"Error when registering user\"}\n",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare mock
			ctrl := gomock.NewController(t)
			f := &fields{
				repo: repository.NewMockRepositoryInterface(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(f)
			}

			// Create a new Echo instance
			e := echo.New()

			// Create a new instance of your server
			s := NewServer(NewServerOptions{Repository: f.repo})

			// Create a request
			req := httptest.NewRequest(http.MethodPost, "/registration", strings.NewReader(tt.args))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Call the handler
			err := s.PostRegistration(c)

			if tt.wantErr {
				assert.Error(t, err)
			}

			// Assert that there is no error
			assert.NoError(t, err)

			// Assert the HTTP status code
			assert.Equal(t, tt.want.httpStatus, rec.Code)

			// Assert the response body
			assert.Equal(t, tt.want.content, rec.Body.String())
		})
	}
}

func TestPostLogin(t *testing.T) {
	// Mock
	type fields struct {
		repo *repository.MockRepositoryInterface
	}

	// Output parameters
	type want struct {
		httpStatus int
		content    string
	}

	// Test Case
	tests := []struct {
		prepare    func(f *fields)
		name       string
		args       string
		want       want
		wantErr    bool
		assertBody bool
	}{
		{
			name: "Success",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, nil)
				f.repo.EXPECT().IncreaseLoginAttempt(gomock.Any(), gomock.Any()).Return(nil)
			},
			args: fmt.Sprintf(`{"phone": "%s", "password": "%s"}`, "+62856712332", "QWErty123!@#"),
			want: want{
				httpStatus: http.StatusOK,
			},
			wantErr:    false,
			assertBody: false,
		}, {
			name: "Invalid request payload",
			prepare: func(f *fields) {

			},
			args: fmt.Sprintf("asd"),
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Invalid request payload\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Invalid phone number",
			prepare: func(f *fields) {

			},
			args: fmt.Sprintf(`{"phone": "%s", "name": "%s", "password": "%s"}`, "123", "User", "Password1!"),
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Invalid phone number. Phone numbers must start with +62 and be 10 to 13 characters in total\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Invalid password",
			prepare: func(f *fields) {

			},
			args: fmt.Sprintf(`{"phone": "%s", "name": "%s", "password": "%s"}`, "+62856712332", "User", ""),
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Invalid password. Passwords must be 6 to 64 characters and contain at least 1 uppercase letter, 1 digit, and 1 special character\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "User not found",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, fmt.Errorf("error"))
			},
			args: fmt.Sprintf(`{"phone": "%s", "name": "%s", "password": "%s"}`, "+62856712332", "User", "Password1!"),
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"User not found\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Invalid password",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, nil)
			},
			args: fmt.Sprintf(`{"phone": "%s", "name": "%s", "password": "%s"}`, "+62856712332", "User", "QWEqwe!@#123"),
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Invalid password\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Failed increase login attempt",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, nil)
				f.repo.EXPECT().IncreaseLoginAttempt(gomock.Any(), gomock.Any()).Return(fmt.Errorf("error"))
			},
			args: fmt.Sprintf(`{"phone": "%s", "password": "%s"}`, "+62856712332", "QWErty123!@#"),
			want: want{
				httpStatus: http.StatusInternalServerError,
				content:    "{\"error\":\"Internal Server Error\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare mock
			ctrl := gomock.NewController(t)
			f := &fields{
				repo: repository.NewMockRepositoryInterface(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(f)
			}

			// Create a new Echo instance
			e := echo.New()

			// Create a new instance of your server
			s := NewServer(NewServerOptions{Repository: f.repo})

			// Create a request
			req := httptest.NewRequest(http.MethodPost, "/registration", strings.NewReader(tt.args))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Call the handler
			err := s.PostLogin(c)

			if tt.wantErr {
				assert.Error(t, err)
			}

			// Assert that there is no error
			assert.NoError(t, err)

			// Assert the HTTP status code
			assert.Equal(t, tt.want.httpStatus, rec.Code)

			if tt.assertBody {
				// Assert the response body
				assert.Equal(t, tt.want.content, rec.Body.String())
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	// Mock
	type fields struct {
		repo *repository.MockRepositoryInterface
	}

	// Output parameters
	type want struct {
		httpStatus int
		content    string
	}

	exp := time.Now().Add(time.Hour * 1)
	token, _ := createToken("123", exp)
	token = "Bearer " + token

	// Test Case
	tests := []struct {
		prepare    func(f *fields)
		name       string
		args       string
		want       want
		wantErr    bool
		assertBody bool
	}{
		{
			name: "Success",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, nil)
			},
			args: token,
			want: want{
				httpStatus: http.StatusOK,
				content:    "{\"name\":\"User\",\"phone\":\"+62856712332\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Forbidden code",
			prepare: func(f *fields) {

			},
			args: "asd",
			want: want{
				httpStatus: http.StatusForbidden,
				content:    "{\"error\":\"Forbidden code\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "User not found",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, fmt.Errorf("error"))
			},
			args: token,
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"User not found\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Authorization token empty",
			prepare: func(f *fields) {

			},
			args: "",
			want: want{
				httpStatus: http.StatusForbidden,
				content:    "{\"error\":\"Forbidden code\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare mock
			ctrl := gomock.NewController(t)
			f := &fields{
				repo: repository.NewMockRepositoryInterface(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(f)
			}

			// Create a new Echo instance
			e := echo.New()

			// Create a new instance of your server
			s := NewServer(NewServerOptions{Repository: f.repo})

			// Create a request
			req := httptest.NewRequest(http.MethodGet, "/profile", strings.NewReader(tt.args))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", tt.args)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Call the handler
			err := s.GetProfile(c)

			if tt.wantErr {
				assert.Error(t, err)
			}

			// Assert that there is no error
			assert.NoError(t, err)

			// Assert the HTTP status code
			assert.Equal(t, tt.want.httpStatus, rec.Code)

			if tt.assertBody {
				// Assert the response body
				assert.Equal(t, tt.want.content, rec.Body.String())
			}
		})
	}
}

func TestPutProfile(t *testing.T) {
	// Mock
	type fields struct {
		repo *repository.MockRepositoryInterface
	}

	// Input parameters
	type args struct {
		jwt     string
		content string
	}

	// Output parameters
	type want struct {
		httpStatus int
		content    string
	}

	exp := time.Now().Add(time.Hour * 1)
	token, _ := createToken("123", exp)

	// Test Case
	tests := []struct {
		prepare    func(f *fields)
		name       string
		args       args
		want       want
		wantErr    bool
		assertBody bool
	}{
		{
			name: "Success",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, nil)
				f.repo.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			args: args{
				jwt:     token,
				content: "{\"name\":\"User\",\"phone\":\"+62856712332\"}",
			},
			want: want{
				httpStatus: http.StatusOK,
				content:    "{\"message\":\"User updated\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Success",
			prepare: func(f *fields) {

			},
			args: args{
				jwt:     "asd",
				content: "{\"name\":\"User\",\"phone\":\"+62856712332\"}",
			},
			want: want{
				httpStatus: http.StatusForbidden,
				content:    "{\"error\":\"Forbidden code\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Success",
			prepare: func(f *fields) {

			},
			args: args{
				jwt:     token,
				content: "asd",
			},
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Invalid request payload\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "User not found",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, fmt.Errorf("User not found"))
			},
			args: args{
				jwt:     token,
				content: "{\"name\":\"User\",\"phone\":\"+62856712332\"}",
			},
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"User not found\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Invalid phone number",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, nil)
			},
			args: args{
				jwt:     token,
				content: "{\"name\":\"User\",\"phone\":\"+122\"}",
			},
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Invalid phone number. Phone numbers must start with +62 and be 10 to 13 characters in total\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Invalid full name",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, nil)
			},
			args: args{
				jwt:     token,
				content: "{\"name\":\"\",\"phone\":\"+62212412422\"}",
			},
			want: want{
				httpStatus: http.StatusBadRequest,
				content:    "{\"error\":\"Invalid full name. Full names must be 3 to 60 characters\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Phone number already exist",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, nil)
				f.repo.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(&pq.Error{Code: "23505"})
			},
			args: args{
				jwt:     token,
				content: "{\"name\":\"User\",\"phone\":\"+62856712332\"}",
			},
			want: want{
				httpStatus: http.StatusConflict,
				content:    "{\"error\":\"Phone number already exist\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Error registering user",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, nil)
				f.repo.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(fmt.Errorf("Error when registering user"))
			},
			args: args{
				jwt:     token,
				content: "{\"name\":\"User\",\"phone\":\"+62856712332\"}",
			},
			want: want{
				httpStatus: http.StatusInternalServerError,
				content:    "{\"error\":\"Error when registering user\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Success without phone",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, nil)
				f.repo.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			args: args{
				jwt:     token,
				content: "{\"name\":\"User\"}",
			},
			want: want{
				httpStatus: http.StatusOK,
				content:    "{\"message\":\"User updated\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		}, {
			name: "Success without name",
			prepare: func(f *fields) {
				f.repo.EXPECT().FindUser(gomock.Any(), gomock.Any()).Return(repository.User{
					ID:       "123",
					Phone:    "+62856712332",
					Name:     "User",
					Password: "$2a$10$Ke5Sl0ra2VeYSmmqjnlE9OLl.I1Bmc8Ou5ix7M2lrPhB6FzV8raJC",
					Salt:     "63RDLuJv8Kmeehqgeg35FA==",
				}, nil)
				f.repo.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			args: args{
				jwt:     token,
				content: "{\"phone\":\"+621234567890\"}",
			},
			want: want{
				httpStatus: http.StatusOK,
				content:    "{\"message\":\"User updated\"}\n",
			},
			wantErr:    false,
			assertBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare mock
			ctrl := gomock.NewController(t)
			f := &fields{
				repo: repository.NewMockRepositoryInterface(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(f)
			}

			// Create a new Echo instance
			e := echo.New()

			// Create a new instance of your server
			s := NewServer(NewServerOptions{Repository: f.repo})

			// Create a request
			req := httptest.NewRequest(http.MethodGet, "/profile", strings.NewReader(tt.args.content))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.args.jwt)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Call the handler
			err := s.PutProfile(c)

			if tt.wantErr {
				assert.Error(t, err)
			}

			// Assert that there is no error
			assert.NoError(t, err)

			// Assert the HTTP status code
			assert.Equal(t, tt.want.httpStatus, rec.Code)

			if tt.assertBody {
				// Assert the response body
				assert.Equal(t, tt.want.content, rec.Body.String())
			}
		})
	}
}

func TestIsValidPhoneNumber(t *testing.T) {
	// Valid phone numbers
	validNumbers := []string{"+62123456789", "+621234567890", "+6212345678901", "+629876543210"}

	// Invalid phone numbers
	invalidNumbers := []string{"123456789", "+6201239", "012345678", "+621234567890525412"}

	for _, num := range validNumbers {
		// Call the isValidPhoneNumber function for valid numbers
		isValid := isValidPhoneNumber(num)

		// Assert that the number is valid
		assert.True(t, isValid, "Expected %s to be a valid phone number", num)
	}

	for _, num := range invalidNumbers {
		// Call the isValidPhoneNumber function for invalid numbers
		isValid := isValidPhoneNumber(num)

		// Assert that the number is invalid
		assert.False(t, isValid, "Expected %s to be an invalid phone number", num)
	}
}

func TestIsValidPassword(t *testing.T) {
	// Valid passwords
	validPasswords := []string{"Abcd123!", "StrongP@ss123", "SecurePwd987!"}

	// Invalid passwords
	invalidPasswords := []string{"short", "weakpassword", "NoSpecialCharacter123", "NoNumber!@#$%^&*()"}

	for _, password := range validPasswords {
		// Call the isValidPassword function for valid passwords
		isValid := isValidPassword(password)

		// Assert that the password is valid
		assert.True(t, isValid, "Expected %s to be a valid password", password)
	}

	for _, password := range invalidPasswords {
		// Call the isValidPassword function for invalid passwords
		isValid := isValidPassword(password)

		// Assert that the password is invalid
		assert.False(t, isValid, "Expected %s to be an invalid password", password)
	}
}

func TestGenerateRandomSalt(t *testing.T) {
	// Call the generateRandomSalt function
	salt, err := generateRandomSalt()

	// Assert that there is no error
	assert.NoError(t, err)

	// Assert that the length of the generated salt is 24 characters (encoded base64 representation of 16 random bytes)
	assert.Equal(t, 24, len(salt))

	// Decode the salt and assert that it is a valid base64 string
	decodedSalt, decodeErr := base64.StdEncoding.DecodeString(salt)
	assert.NoError(t, decodeErr)

	// Assert that the length of the decoded salt is 16 bytes
	assert.Equal(t, 16, len(decodedSalt))
}

func TestHashPassword(t *testing.T) {
	// Set a known password and salt
	password := "MySecurePassword"
	salt := "RandomSalt123"

	// Call the hashPassword function
	hashedPassword, err := hashPassword(password, salt)

	// Assert that there is no error
	assert.NoError(t, err)

	// Compare the hashed password with the expected hash using bcrypt.CompareHashAndPassword
	matchErr := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password+salt))

	// Assert that there is no error in comparing the hashed password
	assert.NoError(t, matchErr, "Expected hashed password to match the original password and salt")
}

func TestCreateToken(t *testing.T) {
	// Set the expiration time to be one hour from now
	expirationTime := time.Now().Add(1 * time.Hour)

	// Call the createToken function
	tokenString, err := createToken("user123", expirationTime)

	// Assert that there is no error
	assert.NoError(t, err)

	// Parse the token to validate its contents
	token, parseErr := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, assert.AnError
		}
		return secret, nil
	})

	// Assert that there is no parsing error
	assert.NoError(t, parseErr)

	// Assert that the token is valid
	assert.True(t, token.Valid)

	// Extract the claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	// Assert that the ID claim matches the expected value
	assert.Equal(t, "user123", claims["id"])

	// Assert that the expiration time claim matches the expected value
	assert.Equal(t, expirationTime.Unix(), int64(claims["exp"].(float64)))
}

func TestValidateToken(t *testing.T) {
	// Create a new Echo instance
	e := echo.New()

	// Set a known user ID
	expectedUserID := "user123"

	// Create a token with the known user ID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  expectedUserID,
		"exp": jwt.TimeFunc().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	t.Run("ValidToken", func(t *testing.T) {
		// Create a request with the token in the Authorization header
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Call the validateToken function
		userID, validateErr := validateToken(c)

		// Assert that there is no error
		assert.NoError(t, validateErr)

		// Assert that the extracted user ID matches the expected user ID
		assert.Equal(t, expectedUserID, userID)
	})

	t.Run("MissingAuthorizationHeader", func(t *testing.T) {
		// Create a request without the Authorization header
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Call the validateToken function
		_, validateErr := validateToken(c)

		// Assert that the error is as expected
		assert.Error(t, validateErr)
		assert.Contains(t, validateErr.Error(), "authorization header is empty")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Create a request with an invalid token in the Authorization header
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Call the validateToken function
		_, validateErr := validateToken(c)

		// Assert that the error is as expected
		assert.Error(t, validateErr)
		assert.Contains(t, validateErr.Error(), "token contains an invalid number of segments")
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		// Create a token with an expiration time in the past
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":  expectedUserID,
			"exp": jwt.TimeFunc().Add(-time.Hour).Unix(),
		})
		expiredTokenString, err := expiredToken.SignedString(secret)
		assert.NoError(t, err)

		// Create a request with the expired token in the Authorization header
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+expiredTokenString)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Call the validateToken function
		_, validateErr := validateToken(c)

		// Assert that the error is as expected
		assert.Error(t, validateErr)
		assert.Contains(t, validateErr.Error(), "Token is expired")
	})
}
