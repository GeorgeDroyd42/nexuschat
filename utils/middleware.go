package utils

import (
	"net/http"

	"auth.com/v4/cache"
	"github.com/labstack/echo/v4"
)

func GetUserID(c echo.Context) string {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}

type AuthMiddlewareConfig struct {
	RedirectOnFailure bool   // Whether to redirect or return error response
	RedirectURL       string // URL to redirect to on authentication failure
	RequireAdmin      bool   // Whether admin privileges are required
	FallbackURL       string // Where to redirect non-admins when RequireAdmin is true
	Return404         bool   // Whether to return 404 instead of redirect or unauthorized
	UseHTTPError      bool   // Whether to use HTTP error instead of JSON response
}

var DefaultAuthMiddlewareConfig = AuthMiddlewareConfig{
	RedirectOnFailure: false,
	RedirectURL:       "/login",
	RequireAdmin:      false,
	FallbackURL:       "/v/main",
}

func GetAuthMiddleware(middlewareType string) echo.MiddlewareFunc {
	switch middlewareType {
	case "api":
		return CreateAuthMiddleware(AuthMiddlewareConfig{
			RedirectOnFailure: false,
			RedirectURL:       "/login",
			RequireAdmin:      false,
		})
	case "user_error":
		return CreateAuthMiddleware(AuthMiddlewareConfig{
			RedirectOnFailure: false,
			RequireAdmin:      false,
			Return404:         false,
			UseHTTPError:      true,
		})

	case "page":
		return CreateAuthMiddleware(AuthMiddlewareConfig{
			RedirectOnFailure: true,
			RedirectURL:       "/login",
			RequireAdmin:      false,
		})
	case "admin_api":
		return CreateAuthMiddleware(AuthMiddlewareConfig{
			RedirectOnFailure: false,
			RequireAdmin:      true,
			Return404:         true,
		})

	case "admin_page":
		return CreateAuthMiddleware(AuthMiddlewareConfig{
			RedirectOnFailure: false,
			RequireAdmin:      true,
			Return404:         true,
		})

	default:
		return CreateAuthMiddleware(DefaultAuthMiddlewareConfig)
	}
}

func RequireParam(c echo.Context, paramName string) (string, error) {
	value := c.Param(paramName)
	if value == "" {
		return "", SendErrorResponse(c, ErrInvalidCredentials)
	}
	return value, nil
}

func RequireUserID(c echo.Context) (string, error) {
	userID := GetUserID(c)
	if userID == "" {
		return "", SendErrorResponse(c, ErrUnauthorized)
	}
	return userID, nil
}
func CreateAuthMiddleware(config AuthMiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, found, err := ValidateUserSession(c)
			if err != nil || !found {
				if config.Return404 {
					return echo.NewHTTPError(http.StatusNotFound)
				}
				if config.RedirectOnFailure {
					return c.Redirect(http.StatusFound, config.RedirectURL)
				}
				return SendErrorResponse(c, ErrUnauthorized)
			}

			cookie, _ := c.Cookie("session")
			sessionData, found, _ := cache.Provider.GetSessionWithUser(cookie.Value)

			if found && sessionData.IsBanned {
				if config.RedirectOnFailure {
					return c.Redirect(http.StatusFound, "/login")
				}
				return SendErrorResponse(c, ErrAccountSuspended)
			} else if !found {
				isBanned, err := IsUserBanned(userID)
				if err == nil && isBanned {
					if config.RedirectOnFailure {
						return c.Redirect(http.StatusFound, "/login")
					}
					return SendErrorResponse(c, ErrAccountSuspended)
				}
			}

			if config.RequireAdmin {
				isAdmin, _ := IsUserAdmin(userID)

				if !isAdmin {
					if config.Return404 {
						return c.Render(http.StatusNotFound, "404.html", nil)
					}
					if config.RedirectOnFailure {
						return c.Redirect(http.StatusFound, config.FallbackURL)
					}
					return SendErrorResponse(c, ErrUnauthorized)
				}
			}

			c.Set("user_id", userID)

			// Sliding window: extend session on every authenticated request
			cookie, err = c.Cookie("session")
			if err == nil {

				GlobalSessionManager.ExtendSessionSafe(cookie.Value)
			}

			return next(c)
		}
	}
}

func GetCSPMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set CSP header before processing request
			cspPolicy := "default-src 'self'; " +
				"script-src 'self'; " +
				"style-src 'self' 'unsafe-inline'; " +
				"img-src 'self' data: blob:; " +
				"font-src 'self'; " +
				"connect-src 'self'; " +
				"media-src 'self'; " +
				"frame-src 'none'; " +
				"object-src 'none'; " +
				"base-uri 'self'"

			c.Response().Header().Set("Content-Security-Policy", cspPolicy)
			return next(c)
		}
	}
}
