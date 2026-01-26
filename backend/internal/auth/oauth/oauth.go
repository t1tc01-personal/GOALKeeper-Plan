package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"goalkeeper-plan/config"
)

var (
	ErrInvalidProvider = errors.New("invalid OAuth provider")
	ErrOAuthExchange   = errors.New("failed to exchange OAuth code")
	ErrOAuthUserInfo   = errors.New("failed to get user info from OAuth provider")
)

type OAuthUser struct {
	ProviderID string
	Email      string
	Name       string
}

type OAuthService struct {
	config config.Configurations
}

func NewOAuthService(cfg config.Configurations) *OAuthService {
	return &OAuthService{
		config: cfg,
	}
}

// generateSecureState generates a cryptographically secure random state string
func generateSecureState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (o *OAuthService) GetAuthURL(provider string) (string, error) {
	state, err := generateSecureState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Get frontend URL, default to localhost:3000 if not set
	frontendURL := o.config.AuthConfig.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	switch provider {
	case "github":
		redirectURI := fmt.Sprintf("%s/auth/github/callback", frontendURL)
		return fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
			o.config.AuthConfig.GitHubClientID,
			url.QueryEscape(redirectURI),
			state,
		), nil
	case "google":
		redirectURI := fmt.Sprintf("%s/auth/google/callback", frontendURL)
		return fmt.Sprintf(
			"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid profile email&state=%s",
			o.config.AuthConfig.GoogleClientID,
			url.QueryEscape(redirectURI),
			state,
		), nil
	default:
		return "", ErrInvalidProvider
	}
}

func (o *OAuthService) ExchangeCode(provider, code string) (*OAuthUser, error) {
	switch provider {
	case "github":
		return o.exchangeGitHubCode(code)
	case "google":
		return o.exchangeGoogleCode(code)
	default:
		return nil, ErrInvalidProvider
	}
}

// GetUserInfoFromToken gets user info from OAuth provider using access token
func (o *OAuthService) GetUserInfoFromToken(provider, accessToken string) (*OAuthUser, error) {
	switch provider {
	case "github":
		return o.getGitHubUserInfo(accessToken)
	case "google":
		return o.getGoogleUserInfo(accessToken)
	default:
		return nil, ErrInvalidProvider
	}
}

func (o *OAuthService) getGitHubUserInfo(accessToken string) (*OAuthUser, error) {
	// Get user info
	userReq, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	userReq.Header.Set("Authorization", "Bearer "+accessToken)
	userReq.Header.Set("Accept", "application/json")

	client := &http.Client{}
	userResp, err := client.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	defer userResp.Body.Close()

	userBody, err := io.ReadAll(userResp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	var githubUser struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.Unmarshal(userBody, &githubUser); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	// If email is not public, get it from emails endpoint
	if githubUser.Email == "" {
		emailReq, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
		}

		emailReq.Header.Set("Authorization", "Bearer "+accessToken)
		emailReq.Header.Set("Accept", "application/json")

		emailResp, err := client.Do(emailReq)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
		}
		defer emailResp.Body.Close()

		emailBody, err := io.ReadAll(emailResp.Body)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
		}

		var emails []struct {
			Email   string `json:"email"`
			Primary bool   `json:"primary"`
		}

		if err := json.Unmarshal(emailBody, &emails); err == nil {
			for _, e := range emails {
				if e.Primary {
					githubUser.Email = e.Email
					break
				}
			}
			if githubUser.Email == "" && len(emails) > 0 {
				githubUser.Email = emails[0].Email
			}
		}
	}

	name := githubUser.Name
	if name == "" {
		name = githubUser.Login
	}

	return &OAuthUser{
		ProviderID: fmt.Sprintf("%d", githubUser.ID),
		Email:      githubUser.Email,
		Name:       name,
	}, nil
}

func (o *OAuthService) getGoogleUserInfo(accessToken string) (*OAuthUser, error) {
	// Get user info
	userReq, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	userReq.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	userResp, err := client.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	defer userResp.Body.Close()

	userBody, err := io.ReadAll(userResp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.Unmarshal(userBody, &googleUser); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	return &OAuthUser{
		ProviderID: googleUser.ID,
		Email:      googleUser.Email,
		Name:       googleUser.Name,
	}, nil
}

func (o *OAuthService) exchangeGitHubCode(code string) (*OAuthUser, error) {
	// Exchange code for access token
	tokenURL := "https://github.com/login/oauth/access_token"
	data := url.Values{}
	data.Set("client_id", o.config.AuthConfig.GitHubClientID)
	data.Set("client_secret", o.config.AuthConfig.GitHubSecret)
	data.Set("code", code)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("%w: %s", ErrOAuthExchange, tokenResp.Error)
	}

	// Get user info
	userReq, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	userReq.Header.Set("Accept", "application/json")

	userResp, err := client.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	defer userResp.Body.Close()

	userBody, err := io.ReadAll(userResp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	var githubUser struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.Unmarshal(userBody, &githubUser); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	// If email is not public, get it from emails endpoint
	if githubUser.Email == "" {
		emailReq, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
		}

		emailReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
		emailReq.Header.Set("Accept", "application/json")

		emailResp, err := client.Do(emailReq)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
		}
		defer emailResp.Body.Close()

		emailBody, err := io.ReadAll(emailResp.Body)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
		}

		var emails []struct {
			Email   string `json:"email"`
			Primary bool   `json:"primary"`
		}

		if err := json.Unmarshal(emailBody, &emails); err == nil {
			for _, e := range emails {
				if e.Primary {
					githubUser.Email = e.Email
					break
				}
			}
			if githubUser.Email == "" && len(emails) > 0 {
				githubUser.Email = emails[0].Email
			}
		}
	}

	name := githubUser.Name
	if name == "" {
		name = githubUser.Login
	}

	return &OAuthUser{
		ProviderID: fmt.Sprintf("%d", githubUser.ID),
		Email:      githubUser.Email,
		Name:       name,
	}, nil
}

func (o *OAuthService) exchangeGoogleCode(code string) (*OAuthUser, error) {
	// Exchange code for access token
	backendURL := o.config.AuthConfig.BackendURL
	if backendURL == "" {
		backendURL = "http://localhost:8000"
	}
	redirectURI := fmt.Sprintf("%s/api/v1/auth/oauth/google/callback", backendURL)
	tokenURL := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("client_id", o.config.AuthConfig.GoogleClientID)
	data.Set("client_secret", o.config.AuthConfig.GoogleSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("%w: %s", ErrOAuthExchange, tokenResp.Error)
	}

	// Get user info
	userReq, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := client.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	defer userResp.Body.Close()

	userBody, err := io.ReadAll(userResp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.Unmarshal(userBody, &googleUser); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	return &OAuthUser{
		ProviderID: googleUser.ID,
		Email:      googleUser.Email,
		Name:       googleUser.Name,
	}, nil
}

