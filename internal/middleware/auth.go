package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/models"
	"github.com/kha/foods-drinks/internal/service"
)

const (
	// AuthorizationHeader is the header key for authorization
	AuthorizationHeader = "Authorization"
	// AuthorizationBearer is the bearer prefix
	AuthorizationBearer = "Bearer"
	// ContextKeyUserID is the context key for user ID
	ContextKeyUserID = "user_id"
	// ContextKeyUserEmail is the context key for user email
	ContextKeyUserEmail = "user_email"
	// ContextKeyUserRole is the context key for user role
	ContextKeyUserRole = "user_role"
	// ContextKeyUser is the context key for user object
	ContextKeyUser = "user"
)

var (
	ErrMissingToken     = errors.New("missing authorization token")
	ErrInvalidToken     = errors.New("invalid or expired token")
	ErrInvalidTokenType = errors.New("invalid token type, expected Bearer")
	ErrUnauthorized     = errors.New("unauthorized access")
	ErrForbidden        = errors.New("forbidden: insufficient permissions")
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth returns a middleware that requires valid JWT token
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.extractAndValidateToken(c)
		if err != nil {
			m.handleAuthError(c, err)
			c.Abort()
			return
		}

		// Set user info in context
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUserEmail, claims.Email)
		c.Set(ContextKeyUserRole, claims.Role)

		c.Next()
	}
}

// RequireAuthWithUser returns a middleware that requires valid JWT token
// and also loads the full user object from database
func (m *AuthMiddleware) RequireAuthWithUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.extractAndValidateToken(c)
		if err != nil {
			m.handleAuthError(c, err)
			c.Abort()
			return
		}

		// Load user from database
		user, err := m.authService.GetUserByID(claims.UserID)
		if err != nil {
			if errors.Is(err, service.ErrUserNotFound) {
				c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
					Error:   "user_not_found",
					Message: "User associated with this token no longer exists",
				})
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to load user information",
			})
			c.Abort()
			return
		}

		// Check user status
		switch user.Status {
		case models.UserStatusInactive:
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "user_inactive",
				Message: "Your account is inactive",
			})
			c.Abort()
			return
		case models.UserStatusBanned:
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "user_banned",
				Message: "Your account has been banned",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set(ContextKeyUserID, user.ID)
		c.Set(ContextKeyUserEmail, user.Email)
		c.Set(ContextKeyUserRole, user.Role)
		c.Set(ContextKeyUser, user)

		c.Next()
	}
}

// RequireAdmin returns a middleware that requires admin role
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyUserRole)
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		if role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "forbidden",
				Message: "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole returns a middleware that requires specific role(s)
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyUserRole)
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "internal_error",
				Message: "Invalid role type in context",
			})
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		allowed := false
		for _, r := range roles {
			if userRole == r {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "forbidden",
				Message: "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth returns a middleware that extracts user info if token is present
// but doesn't fail if token is missing
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.extractAndValidateToken(c)
		if err == nil {
			// Token is valid, set user info in context
			c.Set(ContextKeyUserID, claims.UserID)
			c.Set(ContextKeyUserEmail, claims.Email)
			c.Set(ContextKeyUserRole, claims.Role)
		}
		// Continue regardless of token validity
		c.Next()
	}
}

// extractAndValidateToken extracts and validates JWT token from Authorization header
func (m *AuthMiddleware) extractAndValidateToken(c *gin.Context) (*service.JWTClaims, error) {
	authHeader := c.GetHeader(AuthorizationHeader)
	if authHeader == "" {
		return nil, ErrMissingToken
	}

	// Check Bearer prefix
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != AuthorizationBearer {
		return nil, ErrInvalidTokenType
	}

	tokenString := parts[1]
	if tokenString == "" {
		return nil, ErrMissingToken
	}

	// Validate token
	claims, err := m.authService.ValidateToken(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// handleAuthError handles authentication errors and returns appropriate response
func (m *AuthMiddleware) handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrMissingToken):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "missing_token",
			Message: "Authorization token is required",
		})
	case errors.Is(err, ErrInvalidTokenType):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "invalid_token_type",
			Message: "Invalid token type, expected Bearer token",
		})
	case errors.Is(err, ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "invalid_token",
			Message: "Token is invalid or has expired",
		})
	default:
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication failed",
		})
	}
}

// GetUserID extracts user ID from context
// Returns 0 and false if not found
func GetUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	userID, ok := value.(uint)
	return userID, ok
}

// GetUserEmail extracts user email from context
// Returns empty string and false if not found
func GetUserEmail(c *gin.Context) (string, bool) {
	value, exists := c.Get(ContextKeyUserEmail)
	if !exists {
		return "", false
	}
	email, ok := value.(string)
	return email, ok
}

// GetUserRole extracts user role from context
// Returns empty string and false if not found
func GetUserRole(c *gin.Context) (string, bool) {
	value, exists := c.Get(ContextKeyUserRole)
	if !exists {
		return "", false
	}
	role, ok := value.(string)
	return role, ok
}

// GetUser extracts full user object from context
// Returns nil and false if not found (requires RequireAuthWithUser middleware)
func GetUser(c *gin.Context) (*models.User, bool) {
	value, exists := c.Get(ContextKeyUser)
	if !exists {
		return nil, false
	}
	user, ok := value.(*models.User)
	return user, ok
}

// IsAdmin checks if the current user is an admin
func IsAdmin(c *gin.Context) bool {
	role, ok := GetUserRole(c)
	return ok && role == models.RoleAdmin
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get(ContextKeyUserID)
	return exists
}

// MustGetUserID extracts user ID from context or panics
// Use only when you're certain the user is authenticated
func MustGetUserID(c *gin.Context) uint {
	userID, ok := GetUserID(c)
	if !ok {
		panic("user_id not found in context")
	}
	return userID
}

// MustGetUser extracts user object from context or panics
// Use only when you're certain RequireAuthWithUser middleware was applied
func MustGetUser(c *gin.Context) *models.User {
	user, ok := GetUser(c)
	if !ok {
		panic("user not found in context")
	}
	return user
}
