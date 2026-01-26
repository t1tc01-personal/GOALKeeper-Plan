package service

import (
	"errors"
	"fmt"
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/auth/dto"
	"goalkeeper-plan/internal/auth/jwt"
	"goalkeeper-plan/internal/auth/oauth"
	"goalkeeper-plan/internal/user/model"
	"goalkeeper-plan/internal/user/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserExists            = errors.New("user already exists")
	ErrOAuthFailed           = errors.New("oauth authentication failed")
	ErrOAuthProviderMismatch = errors.New("user already registered with different OAuth provider")
	ErrEmailAlreadyExists    = errors.New("email already exists")
)

type AuthConfiguration func(s *authService) error

type AuthService interface {
	SignUp(req *dto.SignUpRequest) (*dto.AuthResponse, error)
	Login(req *dto.LoginRequest) (*dto.AuthResponse, error)
	Logout(refreshToken string) error
	RefreshToken(refreshToken string) (*dto.AuthResponse, error)
	OAuthLogin(provider, code string) (*dto.AuthResponse, error)
	OAuthLoginWithToken(provider, accessToken string) (*dto.AuthResponse, error)
	GetOAuthURL(provider string) (string, error)
}

type authService struct {
	userRepository repository.UserRepository
	jwtService     *jwt.JWTService
	oauthService   *oauth.OAuthService
	config         config.Configurations
}

func NewAuthService(configs ...AuthConfiguration) (AuthService, error) {
	s := &authService{}
	for _, cfg := range configs {
		if err := cfg(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func WithUserRepository(r repository.UserRepository) AuthConfiguration {
	return func(s *authService) error {
		s.userRepository = r
		return nil
	}
}

func WithJWTService(j *jwt.JWTService) AuthConfiguration {
	return func(s *authService) error {
		s.jwtService = j
		return nil
	}
}

func WithOAuthService(o *oauth.OAuthService) AuthConfiguration {
	return func(s *authService) error {
		s.oauthService = o
		return nil
	}
}

func WithConfig(c config.Configurations) AuthConfiguration {
	return func(s *authService) error {
		s.config = c
		return nil
	}
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPassword compares a password with a hash
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *authService) SignUp(req *dto.SignUpRequest) (*dto.AuthResponse, error) {
	// Check if user already exists
	existingUser, _ := s.userRepository.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: hashedPassword,
	}

	if err := s.userRepository.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserInfo{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}

func (s *authService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepository.GetByEmail(req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user has password (not OAuth-only user)
	if user.PasswordHash == "" {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if !checkPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserInfo{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}

func (s *authService) Logout(refreshToken string) error {
	// In a production system, you might want to blacklist the token in Redis
	// For now, we'll just validate it
	_, err := s.jwtService.ValidateToken(refreshToken)
	return err
}

func (s *authService) RefreshToken(refreshToken string) (*dto.AuthResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userRepository.GetByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Generate new tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User: dto.UserInfo{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}

func (s *authService) GetOAuthURL(provider string) (string, error) {
	return s.oauthService.GetAuthURL(provider)
}

func (s *authService) OAuthLogin(provider, code string) (*dto.AuthResponse, error) {
	// Exchange code for user info
	oauthUser, err := s.oauthService.ExchangeCode(provider, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthFailed, err)
	}

	// Check if user exists by OAuth provider ID (most accurate)
	var user *model.User
	existingUser, _ := s.userRepository.GetByOAuthProviderID(provider, oauthUser.ProviderID)

	if existingUser != nil {
		// User exists with this OAuth provider, use it
		user = existingUser
	} else {
		// Check if user exists by email (might be registered with password or different OAuth provider)
		existingUserByEmail, _ := s.userRepository.GetByEmail(oauthUser.Email)

		if existingUserByEmail != nil {
			// User exists with this email
			if existingUserByEmail.OAuthProvider != "" && existingUserByEmail.OAuthProvider != provider {
				// User already registered with different OAuth provider
				// Allow login but don't overwrite existing OAuth info
				// User can login with any provider if email matches
				user = existingUserByEmail
			} else if existingUserByEmail.OAuthProvider == "" {
				// User registered with password, now linking OAuth provider
				existingUserByEmail.OAuthProvider = provider
				existingUserByEmail.OAuthProviderID = oauthUser.ProviderID
				// Mark email as verified if not already
				if !existingUserByEmail.IsEmailVerified {
					existingUserByEmail.IsEmailVerified = true
				}
				if err := s.userRepository.Update(existingUserByEmail); err != nil {
					return nil, fmt.Errorf("failed to update user: %w", err)
				}
				user = existingUserByEmail
			} else {
				// Same provider but different ID (shouldn't happen, but handle it)
				existingUserByEmail.OAuthProviderID = oauthUser.ProviderID
				// Mark email as verified if not already
				if !existingUserByEmail.IsEmailVerified {
					existingUserByEmail.IsEmailVerified = true
				}
				if err := s.userRepository.Update(existingUserByEmail); err != nil {
					return nil, fmt.Errorf("failed to update user: %w", err)
				}
				user = existingUserByEmail
			}
		} else {
			// User doesn't exist, create new user
			user = &model.User{
				Email:           oauthUser.Email,
				Name:            oauthUser.Name,
				OAuthProvider:   provider,
				OAuthProviderID: oauthUser.ProviderID,
				IsEmailVerified: true, // OAuth emails are typically verified
			}
			if err := s.userRepository.Create(user); err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
		}
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserInfo{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}
