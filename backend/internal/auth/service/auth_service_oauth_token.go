package service

import (
	"fmt"
	"goalkeeper-plan/internal/auth/dto"
	"goalkeeper-plan/internal/user/model"
)

func (s *authService) OAuthLoginWithToken(provider, accessToken string) (*dto.AuthResponse, error) {
	// Get user info from OAuth provider using access token
	oauthUser, err := s.oauthService.GetUserInfoFromToken(provider, accessToken)
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
		// Check if user exists by email
		existingUserByEmail, _ := s.userRepository.GetByEmail(oauthUser.Email)

		if existingUserByEmail != nil {
			// User exists with this email
			if existingUserByEmail.OAuthProvider != "" && existingUserByEmail.OAuthProvider != provider {
				// User already registered with different OAuth provider - return error
				return nil, fmt.Errorf("%w: email already registered with %s", ErrEmailAlreadyExists, existingUserByEmail.OAuthProvider)
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
	accessTokenJWT, err := s.jwtService.GenerateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessTokenJWT,
		RefreshToken: refreshToken,
		User: dto.UserInfo{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}
