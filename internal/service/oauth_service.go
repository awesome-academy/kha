package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kha/foods-drinks/internal/config"
	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/models"
	"github.com/kha/foods-drinks/internal/repository"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var (
	ErrOAuthProviderNotSupported  = errors.New("oauth provider not supported")
	ErrOAuthStateMismatch         = errors.New("oauth state mismatch")
	ErrOAuthCodeExchange          = errors.New("failed to exchange oauth code")
	ErrOAuthUserInfo              = errors.New("failed to get user info from provider")
	ErrOAuthProviderNotConfigured = errors.New("oauth provider not configured")
	ErrOAuthEmailRequired         = errors.New("email is required from oauth provider")
)

// OAuthUserInfo represents user info from OAuth provider
type OAuthUserInfo struct {
	ID            string
	Email         string
	Name          string
	AvatarURL     string
	EmailVerified bool // Whether the email is verified by the provider
}

// OAuthProvider interface for OAuth providers
// This allows easy extension for Facebook, Twitter, etc.
type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error)
	GetProviderName() string
}

// OAuthService handles OAuth authentication
type OAuthService struct {
	userRepo       *repository.UserRepository
	socialAuthRepo *repository.SocialAuthRepository
	authService    *AuthService
	providers      map[string]OAuthProvider
}

// NewOAuthService creates a new OAuthService
func NewOAuthService(
	userRepo *repository.UserRepository,
	socialAuthRepo *repository.SocialAuthRepository,
	authService *AuthService,
	oauthConfig *config.OAuthConfig,
) *OAuthService {
	providers := make(map[string]OAuthProvider)

	// Register Google provider if configured
	if oauthConfig.Google.ClientID != "" && oauthConfig.Google.ClientSecret != "" {
		providers[models.ProviderGoogle] = NewGoogleProvider(&oauthConfig.Google)
	}

	// TODO: Register Facebook provider
	// if oauthConfig.Facebook.ClientID != "" && oauthConfig.Facebook.ClientSecret != "" {
	//     providers[models.ProviderFacebook] = NewFacebookProvider(&oauthConfig.Facebook)
	// }

	// TODO: Register Twitter provider
	// if oauthConfig.Twitter.ClientID != "" && oauthConfig.Twitter.ClientSecret != "" {
	//     providers[models.ProviderTwitter] = NewTwitterProvider(&oauthConfig.Twitter)
	// }

	return &OAuthService{
		userRepo:       userRepo,
		socialAuthRepo: socialAuthRepo,
		authService:    authService,
		providers:      providers,
	}
}

// GetAuthURL returns the OAuth authorization URL for the specified provider
func (s *OAuthService) GetAuthURL(provider, state string) (string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", ErrOAuthProviderNotSupported
	}
	return p.GetAuthURL(state), nil
}

// HandleCallback handles the OAuth callback
func (s *OAuthService) HandleCallback(ctx context.Context, provider, code string) (*dto.AuthResponse, error) {
	p, ok := s.providers[provider]
	if !ok {
		return nil, ErrOAuthProviderNotSupported
	}

	// Exchange code for token
	token, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
	}

	// Get user info from provider
	userInfo, err := p.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	// Find or create user
	user, err := s.findOrCreateUser(userInfo, provider, token)
	if err != nil {
		return nil, err
	}

	// Check user status
	switch user.Status {
	case models.UserStatusInactive:
		return nil, ErrUserInactive
	case models.UserStatusBanned:
		return nil, ErrUserBanned
	}

	// Generate JWT token
	jwtToken, expiresIn, err := s.authService.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken: jwtToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        dto.ToUserResponse(user),
	}, nil
}

// findOrCreateUser finds existing user by social auth or creates new one
func (s *OAuthService) findOrCreateUser(userInfo *OAuthUserInfo, provider string, token *oauth2.Token) (*models.User, error) {
	// Check if social auth exists
	socialAuth, err := s.socialAuthRepo.FindByProviderAndProviderUserID(provider, userInfo.ID)
	if err == nil {
		// Social auth exists, update tokens and return user
		accessToken := token.AccessToken
		var refreshToken *string
		if token.RefreshToken != "" {
			refreshToken = &token.RefreshToken
		}
		if updateErr := s.socialAuthRepo.UpdateTokens(socialAuth.ID, &accessToken, refreshToken); updateErr != nil {
			log.Printf("Failed to update OAuth tokens for social auth ID %d: %v", socialAuth.ID, updateErr)
		}

		user, err := s.userRepo.FindByID(socialAuth.UserID)
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Social auth doesn't exist, check if user with email exists
	var user *models.User
	if userInfo.Email != "" {
		user, err = s.userRepo.FindByEmail(userInfo.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// Use transaction to ensure atomicity when creating user and social auth
	db := s.userRepo.GetDB()
	txErr := db.Transaction(func(tx *gorm.DB) error {
		userRepoTx := s.userRepo.WithTx(tx)
		socialAuthRepoTx := s.socialAuthRepo.WithTx(tx)

		// Create new user if not exists
		if user == nil {
			// Validate that email is not empty before creating user
			if userInfo.Email == "" {
				return ErrOAuthEmailRequired
			}

			user = &models.User{
				Email:     userInfo.Email,
				FullName:  userInfo.Name,
				AvatarURL: &userInfo.AvatarURL,
				Role:      models.RoleUser,
				Status:    models.UserStatusActive,
			}
			// Only set EmailVerifiedAt if provider confirmed the email is verified
			if userInfo.EmailVerified {
				now := time.Now()
				user.EmailVerifiedAt = &now
			}
			if err := userRepoTx.Create(user); err != nil {
				return err
			}
		}

		// Create social auth record
		accessToken := token.AccessToken
		var refreshToken *string
		if token.RefreshToken != "" {
			refreshToken = &token.RefreshToken
		}

		socialAuth = &models.SocialAuth{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: userInfo.ID,
			AccessToken:    &accessToken,
			RefreshToken:   refreshToken,
		}
		if err := socialAuthRepoTx.Create(socialAuth); err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return user, nil
}

// GetSupportedProviders returns list of supported OAuth providers
func (s *OAuthService) GetSupportedProviders() []string {
	providers := make([]string, 0, len(s.providers))
	for name := range s.providers {
		providers = append(providers, name)
	}
	return providers
}

// IsProviderSupported checks if a provider is supported
func (s *OAuthService) IsProviderSupported(provider string) bool {
	_, ok := s.providers[provider]
	return ok
}

// ========================================
// Google OAuth Provider Implementation
// ========================================

// GoogleProvider implements OAuthProvider for Google
type GoogleProvider struct {
	config *oauth2.Config
}

// NewGoogleProvider creates a new GoogleProvider
func NewGoogleProvider(cfg *config.OAuthProviderConfig) *GoogleProvider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes: []string{
				"openid", // Required for OpenID Connect compliance
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

// GetAuthURL returns the Google OAuth authorization URL
func (p *GoogleProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges the authorization code for a token
func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

// GetUserInfo retrieves user information from Google
func (p *GoogleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:            googleUser.ID,
		Email:         strings.ToLower(strings.TrimSpace(googleUser.Email)), // Normalize email
		Name:          googleUser.Name,
		AvatarURL:     googleUser.Picture,
		EmailVerified: googleUser.VerifiedEmail,
	}, nil
}

// GetProviderName returns the provider name
func (p *GoogleProvider) GetProviderName() string {
	return models.ProviderGoogle
}

// ========================================
// Facebook OAuth Provider (Placeholder)
// ========================================

// TODO: Implement FacebookProvider
// type FacebookProvider struct {
//     config *oauth2.Config
// }
//
// func NewFacebookProvider(cfg *config.OAuthProviderConfig) *FacebookProvider {
//     return &FacebookProvider{
//         config: &oauth2.Config{
//             ClientID:     cfg.ClientID,
//             ClientSecret: cfg.ClientSecret,
//             RedirectURL:  cfg.RedirectURL,
//             Scopes:       []string{"email", "public_profile"},
//             Endpoint:     facebook.Endpoint,
//         },
//     }
// }

// ========================================
// Twitter OAuth Provider (Placeholder)
// ========================================

// TODO: Implement TwitterProvider (OAuth 2.0)
// Note: Twitter uses OAuth 2.0 with PKCE for new apps
// type TwitterProvider struct {
//     config *oauth2.Config
// }
