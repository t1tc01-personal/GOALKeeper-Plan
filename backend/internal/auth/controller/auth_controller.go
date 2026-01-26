package controller

import (
	"context"
	"fmt"
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/api"
	"goalkeeper-plan/internal/auth/dto"
	authMessages "goalkeeper-plan/internal/auth/messages"
	"goalkeeper-plan/internal/auth/service"
	tokenStorage "goalkeeper-plan/internal/auth/token_storage"
	"goalkeeper-plan/internal/errors"
	"goalkeeper-plan/internal/logger"
	userMessages "goalkeeper-plan/internal/user/messages"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthController interface {
	SignUp(*gin.Context)
	Login(*gin.Context)
	Logout(*gin.Context)
	RefreshToken(*gin.Context)
	GetOAuthURL(*gin.Context)
	OAuthCallback(*gin.Context)
	OAuthCallbackGET(*gin.Context)
	OAuthTokenCallback(*gin.Context)
	GetOAuthTokens(*gin.Context)
}

type authController struct {
	authService  service.AuthService
	tokenStorage *tokenStorage.TokenStorage
	config       config.Configurations
	logger       logger.Logger
}

func NewAuthController(s service.AuthService, ts *tokenStorage.TokenStorage, cfg config.Configurations, l logger.Logger) AuthController {
	return &authController{
		authService:  s,
		tokenStorage: ts,
		config:       cfg,
		logger:       l,
	}
}

// SignUp godoc
// @Summary      Sign up a new user
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request   body      dto.SignUpRequest  true  "Sign up request"
// @Success      201       {object}  response.Response{data=dto.AuthResponse}
// @Failure      400       {object}  response.Response
// @Failure      409       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /auth/signup [post]
func (c *authController) SignUp(ctx *gin.Context) {
	api.HandleRequestWithStatus(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, 201, authMessages.MsgSignUpSuccess, func(ctx *gin.Context, req dto.SignUpRequest) (interface{}, error) {
		authResponse, err := c.authService.SignUp(&req)
		if err != nil {
			if err == service.ErrUserExists {
				c.logger.Error("User already exists", zap.String("email", req.Email))
				return nil, errors.NewConflictError(
					errors.CodeUserExists,
					userMessages.MsgUserExists,
					err,
				)
			}
			c.logger.Error("Failed to sign up", zap.Error(err), zap.String("email", req.Email))
			return nil, errors.NewInternalError(
				errors.CodeFailedToSignUp,
				authMessages.MsgFailedToSignUp,
				err,
			)
		}
		return authResponse, nil
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request   body      dto.LoginRequest  true  "Login request"
// @Success      200       {object}  response.Response{data=dto.AuthResponse}
// @Failure      400       {object}  response.Response
// @Failure      401       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /auth/login [post]
func (c *authController) Login(ctx *gin.Context) {
	api.HandleRequestWithStatus(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, 200, authMessages.MsgLoginSuccess, func(ctx *gin.Context, req dto.LoginRequest) (interface{}, error) {
		authResponse, err := c.authService.Login(&req)
		if err != nil {
			if err == service.ErrInvalidCredentials {
				c.logger.Error("Invalid credentials", zap.String("email", req.Email))
				return nil, errors.NewAuthenticationError(
					errors.CodeInvalidCredentials,
					authMessages.MsgInvalidCredentials,
					err,
				)
			}
			c.logger.Error("Failed to login", zap.Error(err), zap.String("email", req.Email))
			return nil, errors.NewInternalError(
				errors.CodeFailedToLogin,
				authMessages.MsgFailedToLogin,
				err,
			)
		}
		return authResponse, nil
	})
}

// Logout godoc
// @Summary      Logout user
// @Description  Logout user by invalidating refresh token and clearing cookies
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request   body      dto.LogoutRequest  true  "Logout request"
// @Success      200       {object}  response.Response
// @Failure      400       {object}  response.Response
// @Router       /auth/logout [post]
func (c *authController) Logout(ctx *gin.Context) {
	api.HandleRequestWithStatus(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, 200, authMessages.MsgLogoutSuccess, func(ctx *gin.Context, req dto.LogoutRequest) (interface{}, error) {
		if err := c.authService.Logout(req.RefreshToken); err != nil {
			c.logger.Error("Failed to logout", zap.Error(err))
			return nil, errors.NewValidationError(
				errors.CodeInvalidRefreshToken,
				authMessages.MsgInvalidRefreshToken,
				err,
			)
		}

		// Clear access token cookie
		// Set maxAge to -1 to delete the cookie, match path and domain with frontend
		ctx.SetCookie("accessToken", "", -1, "/", "", false, false)
		// Clear refresh token cookie
		ctx.SetCookie("refreshToken", "", -1, "/", "", false, false)

		return nil, nil
	})
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Get new access and refresh tokens using refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request   body      dto.RefreshTokenRequest  true  "Refresh token request"
// @Success      200       {object}  response.Response{data=dto.AuthResponse}
// @Failure      400       {object}  response.Response
// @Failure      401       {object}  response.Response
// @Router       /auth/refresh [post]
func (c *authController) RefreshToken(ctx *gin.Context) {
	api.HandleRequestWithStatus(ctx, api.HandlerConfig{
		Logger: c.logger,
	}, 200, authMessages.MsgTokenRefreshed, func(ctx *gin.Context, req dto.RefreshTokenRequest) (interface{}, error) {
		authResponse, err := c.authService.RefreshToken(req.RefreshToken)
		if err != nil {
			c.logger.Error("Failed to refresh token", zap.Error(err))
			return nil, errors.NewAuthenticationError(
				errors.CodeTokenExpired,
				authMessages.MsgExpiredRefreshToken,
				err,
			)
		}
		return authResponse, nil
	})
}

// GetOAuthURL godoc
// @Summary      Get OAuth authorization URL
// @Description  Get OAuth authorization URL for GitHub or Google
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        provider   path      string  true  "OAuth provider (github or google)"
// @Success      200        {object}  response.Response{data=object}
// @Failure      400        {object}  response.Response
// @Router       /auth/oauth/{provider}/url [get]
func (c *authController) GetOAuthURL(ctx *gin.Context) {
	api.HandleParamRequestWithMessage(ctx, "provider", api.HandlerConfig{
		Logger: c.logger,
	}, authMessages.MsgOAuthURLGenerated, func(ctx *gin.Context, provider string) (interface{}, error) {
		url, err := c.authService.GetOAuthURL(provider)
		if err != nil {
			c.logger.Error("Failed to get OAuth URL", zap.Error(err), zap.String("provider", provider))
			return nil, errors.NewValidationError(
				errors.CodeInvalidOAuthProvider,
				authMessages.MsgInvalidOAuthProvider,
				err,
			)
		}
		return gin.H{"url": url}, nil
	})
}

// OAuthCallback godoc
// @Summary      OAuth callback
// @Description  Handle OAuth callback and authenticate user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        provider   path      string                  true  "OAuth provider (github or google)"
// @Param        request    body      dto.OAuthCallbackRequest  true  "OAuth callback request"
// @Success      200        {object}  response.Response{data=dto.AuthResponse}
// @Failure      400        {object}  response.Response
// @Failure      401        {object}  response.Response
// @Router       /auth/oauth/{provider}/callback [post]
func (c *authController) OAuthCallback(ctx *gin.Context) {
	provider, ok := api.GetParam(ctx, "provider", authMessages.MsgOAuthProviderRequired)
	if !ok {
		return
	}

	req, ok := api.BindRequest[dto.OAuthCallbackRequest](ctx, c.logger)
	if !ok {
		return
	}

	authResponse, err := c.authService.OAuthLogin(provider, req.Code)
	if err != nil {
		c.logger.Error("Failed to process OAuth callback", zap.Error(err), zap.String("provider", provider))

		// Check if it's email already exists error
		if err == service.ErrEmailAlreadyExists ||
			strings.Contains(err.Error(), "email already registered") {
			api.HandleError(ctx, errors.NewConflictError(
				errors.CodeOAuthEmailExists,
				authMessages.MsgOAuthEmailExists,
				err,
			), c.logger)
			return
		}

		api.HandleError(ctx, errors.NewAuthenticationError(
			errors.CodeOAuthFailed,
			authMessages.MsgOAuthFailed,
			err,
		), c.logger)
		return
	}

	api.SendSuccess(ctx, 200, authMessages.MsgOAuthSuccess, authResponse)
}

// OAuthCallbackGET godoc
// @Summary      OAuth callback (GET redirect)
// @Description  Handle OAuth redirect callback from provider, store tokens and redirect to frontend
// @Tags         auth
// @Param        provider   path      string  true  "OAuth provider (github or google)"
// @Param        code       query     string  true  "Authorization code"
// @Param        state      query     string  false "State parameter"
// @Param        error      query     string  false "Error from OAuth provider"
// @Success      302        Redirect to frontend
// @Failure      400        Redirect to frontend with error
// @Failure      401        Redirect to frontend with error
// @Router       /auth/oauth/{provider}/callback [get]
func (c *authController) OAuthCallbackGET(ctx *gin.Context) {
	provider, ok := api.GetParam(ctx, "provider", authMessages.MsgOAuthProviderRequired)
	if !ok {
		c.redirectToFrontendWithError(ctx, provider, "invalid_provider")
		return
	}

	// Check for OAuth error
	if oauthError := ctx.Query("error"); oauthError != "" {
		c.logger.Error("OAuth provider returned error", zap.String("provider", provider), zap.String("error", oauthError))
		c.redirectToFrontendWithError(ctx, provider, oauthError)
		return
	}

	code := ctx.Query("code")
	if code == "" {
		c.logger.Error("Missing authorization code", zap.String("provider", provider))
		c.redirectToFrontendWithError(ctx, provider, "missing_code")
		return
	}

	// Exchange code for tokens
	authResponse, err := c.authService.OAuthLogin(provider, code)
	if err != nil {
		c.logger.Error("Failed to process OAuth callback", zap.Error(err), zap.String("provider", provider))

		// Check if it's email already exists error
		if err == service.ErrEmailAlreadyExists ||
			strings.Contains(err.Error(), "email already registered") {
			c.redirectToFrontendWithError(ctx, provider, "email_already_exists")
			return
		}

		c.redirectToFrontendWithError(ctx, provider, "oauth_failed")
		return
	}

	// Store tokens in Redis
	tokenData := &tokenStorage.TokenData{
		AccessToken:  authResponse.AccessToken,
		RefreshToken: authResponse.RefreshToken,
		UserID:       authResponse.User.ID,
		Email:        authResponse.User.Email,
		Name:         authResponse.User.Name,
	}

	tokenID, err := c.tokenStorage.StoreToken(context.Background(), tokenData)
	if err != nil {
		c.logger.Error("Failed to store tokens", zap.Error(err))
		c.redirectToFrontendWithError(ctx, provider, "storage_failed")
		return
	}

	// Redirect to frontend with token ID
	frontendURL := c.config.AuthConfig.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	redirectURL := fmt.Sprintf("%s/auth/oauth/callback/%s?token_id=%s", frontendURL, provider, tokenID)
	ctx.Redirect(http.StatusFound, redirectURL)
}

// GetOAuthTokens godoc
// @Summary      Get OAuth tokens by token ID
// @Description  Retrieve OAuth tokens using token ID (one-time use, tokens are deleted after retrieval)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        token_id   path      string  true  "Token ID from OAuth callback"
// @Success      200        {object}  response.Response{data=tokenStorage.TokenData}
// @Failure      400        {object}  response.Response
// @Failure      404        {object}  response.Response
// @Router       /auth/oauth/tokens/{token_id} [get]
func (c *authController) GetOAuthTokens(ctx *gin.Context) {
	tokenID, ok := api.GetParam(ctx, "token_id", "Token ID is required")
	if !ok {
		return
	}

	// Get and delete token (one-time use)
	tokenData, err := c.tokenStorage.GetAndDeleteToken(context.Background(), tokenID)
	if err != nil {
		c.logger.Error("Failed to get tokens", zap.Error(err), zap.String("token_id", tokenID))
		api.HandleError(ctx, errors.NewNotFoundError(
			errors.CodeTokenInvalid,
			"Token not found or expired",
			err,
		), c.logger)
		return
	}

	// Return tokens
	api.SendSuccess(ctx, 200, "", tokenData)
}

// OAuthTokenCallback godoc
// @Summary      OAuth callback with token
// @Description  Handle OAuth callback with access token from provider
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        provider   path      string                        true  "OAuth provider (github or google)"
// @Param        request    body      dto.OAuthTokenCallbackRequest  true  "OAuth token callback request"
// @Success      200        {object}  response.Response{data=dto.AuthResponse}
// @Failure      400        {object}  response.Response
// @Failure      401        {object}  response.Response
// @Failure      409        {object}  response.Response
// @Router       /auth/oauth/{provider}/token-callback [post]
func (c *authController) OAuthTokenCallback(ctx *gin.Context) {
	provider, ok := api.GetParam(ctx, "provider", authMessages.MsgOAuthProviderRequired)
	if !ok {
		return
	}

	req, ok := api.BindRequest[dto.OAuthTokenCallbackRequest](ctx, c.logger)
	if !ok {
		return
	}

	authResponse, err := c.authService.OAuthLoginWithToken(provider, req.AccessToken)
	if err != nil {
		c.logger.Error("Failed to process OAuth token callback", zap.Error(err), zap.String("provider", provider))

		// Check if it's email already exists error
		if err == service.ErrEmailAlreadyExists ||
			strings.Contains(err.Error(), "email already registered") {
			api.HandleError(ctx, errors.NewConflictError(
				errors.CodeOAuthEmailExists,
				authMessages.MsgOAuthEmailExists,
				err,
			), c.logger)
			return
		}

		api.HandleError(ctx, errors.NewAuthenticationError(
			errors.CodeOAuthFailed,
			authMessages.MsgOAuthFailed,
			err,
		), c.logger)
		return
	}

	api.SendSuccess(ctx, 200, authMessages.MsgOAuthSuccess, authResponse)
}

// redirectToFrontendWithError redirects to frontend with error parameter
func (c *authController) redirectToFrontendWithError(ctx *gin.Context, provider, errorMsg string) {
	frontendURL := c.config.AuthConfig.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	redirectURL := fmt.Sprintf("%s/auth/oauth/callback/%s?error=%s", frontendURL, provider, errorMsg)
	ctx.Redirect(http.StatusFound, redirectURL)
}
