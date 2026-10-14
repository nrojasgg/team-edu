package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"ms-registro-curricular/internal/dto"
	"ms-registro-curricular/internal/models"
	"ms-registro-curricular/internal/repositories"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthService interface {
	Register(user *models.User) (*models.User, error)
	Login(email, password string) (string, *models.User, error)
	GoogleLogin(code string) (string, *models.User, error)
	GetGoogleAuthURL() string
	ValidateToken(tokenString string) (*jwt.Token, error)
	GenerateToken(userID int64) (string, error)
	ForgotPassword(email string, resetURL string) error
	ResetPassword(token string, newPassword string) error
}

type authService struct {
	userRepo      repositories.UserRepository
	roleRepo      repositories.RoleRepository
	jwtSecret     string
	googleConfig  *oauth2.Config
	apiGatewayURL string
}

func NewAuthService(userRepo repositories.UserRepository, roleRepo repositories.RoleRepository, jwtSecret, googleClientID, googleClientSecret, googleRedirectURL, apiGatewayURL string) AuthService {
	googleConfig := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  googleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &authService{
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		jwtSecret:     jwtSecret,
		googleConfig:  googleConfig,
		apiGatewayURL: apiGatewayURL,
	}
}

func (s *authService) Register(user *models.User) (*models.User, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(user.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	existingUser, _ = s.userRepo.GetByRUT(user.RUT)
	if existingUser != nil {
		return nil, errors.New("user with this RUT already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	user.Password = string(hashedPassword)

	// Create user
	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	// Get user with role information
	createdUser, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		return nil, err
	}

	// Clear password before returning
	createdUser.Password = ""
	return createdUser, nil
}

func (s *authService) Login(email, password string) (string, *models.User, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", nil, errors.New("Correo o contraseña incorrectos")
	}

	// Check if user has a password set (Google OAuth users may not have one)
	if user.Password == "" {
		return "", nil, errors.New("this account uses Google login. Please use 'Sign in with Google' or set a password via 'Forgot Password'")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, errors.New("Correo o contraseña incorrectos")
	}

	// Generate token
	token, err := s.GenerateToken(user.ID)
	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	// Clear password before returning
	user.Password = ""
	return token, user, nil
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

func (s *authService) GenerateToken(userID int64) (string, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user for token: %w", err)
	}

	permissions, err := s.userRepo.GetUserPermissions(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get permissions for token: %w", err)
	}

	permNames := make([]string, 0, len(permissions))
	for _, p := range permissions {
		permNames = append(permNames, p.Name)
	}

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	claims := jwt.MapClaims{
		"id":          userID,
		"role":        roleName,
		"permissions": permNames,
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) GetGoogleAuthURL() string {
	return s.googleConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

func (s *authService) GoogleLogin(code string) (string, *models.User, error) {
	fmt.Printf("GoogleLogin called with code: %s\n", code)

	// Exchange code for token
	token, err := s.googleConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("Error exchanging code for token: %v\n", err)
		return "", nil, fmt.Errorf("failed to exchange code for token: %v", err)
	}

	fmt.Printf("Successfully exchanged code for token\n")

	// Get user info from Google
	userInfo, err := s.getGoogleUserInfo(token.AccessToken)
	if err != nil {
		fmt.Printf("Error getting user info: %v\n", err)
		return "", nil, fmt.Errorf("failed to get user info: %v", err)
	}

	fmt.Printf("Got user info: Name=%s, Email=%s, Picture=%s\n", userInfo.Name, userInfo.Email, userInfo.Picture)

	// Validate email domain
	if !s.isValidUsachEmail(userInfo.Email) {
		fmt.Printf("Invalid email domain: %s\n", userInfo.Email)
		return "", nil, errors.New("only @usach.cl email addresses are allowed")
	}

	fmt.Printf("Email domain validation passed\n")

	// Check if user exists
	user, err := s.userRepo.GetByEmail(userInfo.Email)
	if err != nil {
		fmt.Printf("User doesn't exist, creating new user\n")

		// Get the "estudiante" role by name
		estudianteRole, err := s.roleRepo.GetByName("estudiante")
		if err != nil {
			fmt.Printf("Error finding estudiante role: %v\n", err)
			return "", nil, fmt.Errorf("failed to find estudiante role: %v", err)
		}

		// User doesn't exist, create new user with estudiante role
		// Separar el nombre de Google en nombres y apellidos
		nameParts := strings.Fields(userInfo.Name)
		nombres := ""
		paterno := ""
		materno := ""

		if len(nameParts) > 0 {
			nombres = nameParts[0]
		}
		if len(nameParts) > 1 {
			paterno = nameParts[1]
		}
		if len(nameParts) > 2 {
			materno = strings.Join(nameParts[2:], " ")
		}

		user = &models.User{
			Nombres:    nombres,
			Paterno:    paterno,
			Materno:    materno,
			Email:      userInfo.Email,
			PictureURL: userInfo.Picture,
			RoleID:     estudianteRole.ID, // Use the actual ID of "estudiante" role
			CareerCode: nil,               // No career code for Google OAuth users by default
			Password:   "",                // No password needed for Google OAuth users
		}

		err = s.userRepo.Create(user)
		if err != nil {
			fmt.Printf("Error creating user: %v\n", err)
			return "", nil, fmt.Errorf("failed to create user: %v", err)
		}

		fmt.Printf("User created successfully with ID: %d and role: %s\n", user.ID, estudianteRole.Name)

		// Get user with role information
		user, err = s.userRepo.GetByID(user.ID)
		if err != nil {
			fmt.Printf("Error getting created user: %v\n", err)
			return "", nil, fmt.Errorf("failed to get created user: %v", err)
		}
	} else {
		fmt.Printf("User exists, updating picture URL if needed\n")
		// User exists, update picture URL if it has changed
		if user.PictureURL != userInfo.Picture {
			user.PictureURL = userInfo.Picture
			err = s.userRepo.Update(user)
			if err != nil {
				// Log error but don't fail the login
				fmt.Printf("Warning: failed to update user picture URL: %v\n", err)
			}
		}
	}

	fmt.Printf("User ready: ID=%d, Name=%s %s %s, Email=%s\n", user.ID, user.Nombres, user.Paterno, user.Materno, user.Email)

	// Generate JWT token
	jwtToken, err := s.GenerateToken(user.ID)
	if err != nil {
		fmt.Printf("Error generating JWT token: %v\n", err)
		return "", nil, fmt.Errorf("failed to generate token: %v", err)
	}

	fmt.Printf("JWT token generated successfully\n")

	// Clear password before returning
	user.Password = ""
	return jwtToken, user, nil
}

func (s *authService) getGoogleUserInfo(accessToken string) (*dto.GoogleUserInfo, error) {
	url := "https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo dto.GoogleUserInfo
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s *authService) isValidUsachEmail(email string) bool {
	return strings.HasSuffix(strings.ToLower(email), "@usach.cl")
}

func (s *authService) ForgotPassword(email string, resetURL string) error {
	// 1. Verificar si el usuario existe
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Por seguridad, no revelamos si el correo existe o no
		return nil
	}

	// 2. Generar token de reseteo corto (1 hora)
	claims := jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour).Unix(),
		"type":  "password_reset",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return fmt.Errorf("error generating reset token: %w", err)
	}

	// 3. Llamar a ms-notis a través del API Gateway
	notificationData := map[string]interface{}{
		"email":       user.Email,
		"name":        fmt.Sprintf("%s %s", user.Nombres, user.Paterno),
		"reset_token": tokenString,
		"reset_url":   resetURL,
	}
	jsonData, err := json.Marshal(notificationData)
	if err != nil {
		return fmt.Errorf("error encoding notification data: %w", err)
	}

	notisURL := fmt.Sprintf("%s/api/v1/notifications/password-reset", s.apiGatewayURL)
	log.Printf("Calling notification service at: %s", notisURL)
	resp, err := http.Post(notisURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("Error calling notification service: %v", err)
		return fmt.Errorf("error calling notification service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Notification service returned error %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("notification service returned error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *authService) ResetPassword(tokenString string, newPassword string) error {
	// 1. Validar token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return errors.New("invalid or expired reset token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "password_reset" {
		return errors.New("invalid token type")
	}

	userID := int64(claims["id"].(float64))

	// 2. Obtener usuario
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// 3. Hashear nueva contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	// 4. Actualizar usuario
	user.Password = string(hashedPassword)
	return s.userRepo.Update(user)
}
