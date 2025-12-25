package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"avatar-face-swap-go/internal/config"
)

var (
	ErrKeycloakNotConfigured = errors.New("keycloak is not configured")
	ErrTokenExchangeFailed   = errors.New("failed to exchange token")
	ErrDiscoveryFailed       = errors.New("failed to fetch OIDC discovery")
)

// OIDCDiscovery represents the OpenID Connect discovery document
type OIDCDiscovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
}

// OIDCTokenResponse represents the OAuth2 token response from Keycloak
type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// OIDCUserInfo represents user info from Keycloak userinfo endpoint
type OIDCUserInfo struct {
	Sub               string `json:"sub"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
}

// KeycloakService handles Keycloak OIDC authentication
type KeycloakService struct {
	config     *config.Config
	discovery  *OIDCDiscovery
	httpClient *http.Client
	mu         sync.RWMutex
}

var (
	keycloakInstance *KeycloakService
	keycloakOnce     sync.Once
)

// GetKeycloakService returns the singleton KeycloakService instance
func GetKeycloakService() *KeycloakService {
	keycloakOnce.Do(func() {
		keycloakInstance = &KeycloakService{
			config: config.Load(),
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}
	})
	return keycloakInstance
}

// IsEnabled returns true if Keycloak is configured
func (k *KeycloakService) IsEnabled() bool {
	return k.config.IsKeycloakEnabled()
}

// GetClientID returns the Keycloak client ID
func (k *KeycloakService) GetClientID() string {
	return k.config.KeycloakClientID
}

// fetchDiscovery fetches and caches the OIDC discovery document
func (k *KeycloakService) fetchDiscovery(ctx context.Context) (*OIDCDiscovery, error) {
	k.mu.RLock()
	if k.discovery != nil {
		defer k.mu.RUnlock()
		return k.discovery, nil
	}
	k.mu.RUnlock()

	// Need to fetch discovery
	k.mu.Lock()
	defer k.mu.Unlock()

	// Double check after acquiring write lock
	if k.discovery != nil {
		return k.discovery, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.config.KeycloakServerURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDiscoveryFailed, err)
	}

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDiscoveryFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrDiscoveryFailed, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDiscoveryFailed, err)
	}

	var discovery OIDCDiscovery
	if err := json.Unmarshal(body, &discovery); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDiscoveryFailed, err)
	}

	k.discovery = &discovery
	return k.discovery, nil
}

// GetDiscovery returns the cached OIDC discovery document
func (k *KeycloakService) GetDiscovery(ctx context.Context) (*OIDCDiscovery, error) {
	if !k.IsEnabled() {
		return nil, ErrKeycloakNotConfigured
	}
	return k.fetchDiscovery(ctx)
}

// GetAuthorizationURL returns the URL to redirect the user to for authentication
func (k *KeycloakService) GetAuthorizationURL(ctx context.Context, redirectURI, state string) (string, error) {
	discovery, err := k.GetDiscovery(ctx)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Set("client_id", k.config.KeycloakClientID)
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("redirect_uri", redirectURI)
	if state != "" {
		params.Set("state", state)
	}

	return discovery.AuthorizationEndpoint + "?" + params.Encode(), nil
}

// ExchangeCode exchanges an authorization code for tokens
func (k *KeycloakService) ExchangeCode(ctx context.Context, code, redirectURI string) (*OIDCTokenResponse, error) {
	discovery, err := k.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", k.config.KeycloakClientID)
	data.Set("client_secret", k.config.KeycloakClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, discovery.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrTokenExchangeFailed, string(body))
	}

	var tokenResp OIDCTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo fetches user info using an access token
func (k *KeycloakService) GetUserInfo(ctx context.Context, accessToken string) (*OIDCUserInfo, error) {
	discovery, err := k.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discovery.UserInfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read userinfo response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo failed: %s", string(body))
	}

	var userInfo OIDCUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse userinfo response: %w", err)
	}

	return &userInfo, nil
}

// GetLogoutURL returns the URL to redirect the user to for logout
func (k *KeycloakService) GetLogoutURL(ctx context.Context, postLogoutRedirectURI string) (string, error) {
	discovery, err := k.GetDiscovery(ctx)
	if err != nil {
		return "", err
	}

	// Build logout URL
	logoutURL := discovery.EndSessionEndpoint
	if logoutURL == "" {
		// Fallback for older Keycloak versions
		logoutURL = discovery.Issuer + "/protocol/openid-connect/logout"
	}

	params := url.Values{}
	if postLogoutRedirectURI != "" {
		params.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	}
	params.Set("client_id", k.config.KeycloakClientID)

	return logoutURL + "?" + params.Encode(), nil
}

// GetFrontendBaseURL returns the frontend base URL for redirects
func (k *KeycloakService) GetFrontendBaseURL() string {
	return k.config.FrontendBaseURL
}
