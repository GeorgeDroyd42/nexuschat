package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"auth.com/v4/api"
	"auth.com/v4/cache"
	"auth.com/v4/internal/csrf"
	"auth.com/v4/internal/invite"
	"auth.com/v4/internal/webhook"
	"auth.com/v4/utils"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	AUTH_NONE  = 0 // No authentication required
	AUTH_USER  = 1 // User authentication required
	AUTH_ADMIN = 2 // Admin authentication required
)

type Route struct {
	Method    string
	Path      string
	Handler   echo.HandlerFunc
	AuthLevel int
}

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func RegisterRoutes(e *echo.Echo, routes []Route, useCSRF bool) {
	// Create CSRF middleware once
	csrfMiddleware := middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token",
		ContextKey:     "csrf",
		CookieName:     "csrf_token",
		CookiePath:     "/",
		CookieMaxAge:   3600,
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
	})

	for _, r := range routes {
		middlewares := []echo.MiddlewareFunc{}

		if useCSRF {
			middlewares = append(middlewares, csrfMiddleware)
		}

		// Add auth middleware if needed
		switch r.AuthLevel {
		case AUTH_USER:
			if len(r.Path) >= 5 && r.Path[:5] == "/api/" {
				middlewares = append(middlewares, utils.GetAuthMiddleware("api"))
			} else if len(r.Path) >= 3 && r.Path[:3] == "/i/" {
				middlewares = append(middlewares, utils.GetAuthMiddleware("user_error"))
			} else {
				middlewares = append(middlewares, utils.GetAuthMiddleware("page"))
			}

		case AUTH_ADMIN:
			if len(r.Path) >= 5 && r.Path[:5] == "/api/" {
				middlewares = append(middlewares, utils.GetAuthMiddleware("admin_api"))
			} else {
				middlewares = append(middlewares, utils.GetAuthMiddleware("admin_page"))
			}
		}

		// Register route with all middlewares
		switch r.Method {
		case "GET":
			e.GET(r.Path, r.Handler, middlewares...)
		case "POST":
			e.POST(r.Path, r.Handler, middlewares...)
		case "PUT":
			e.PUT(r.Path, r.Handler, middlewares...)
		case "DELETE":
			e.DELETE(r.Path, r.Handler, middlewares...)
		}
	}
}

func init() {
	utils.RegisterMigration(utils.Migration{
		Name: "change_user_id_to_varchar",
		Up: func(db *sql.DB) error {
			return utils.AddIfNotExists("user_id")
		},
	})
	utils.RegisterMigration(utils.Migration{
		Name: "change_sessions_username_to_user_id",
		Up: func(db *sql.DB) error {
			return utils.AddIfNotExists("user_id")
		},
	})

	utils.RegisterMigration(utils.Migration{
		Name: "add_is_banned_column",
		Up: func(db *sql.DB) error {
			return utils.AddIfNotExists("is_banned")
		},
	})

}

func startRedisSubscriber() {
	go func() {
		for {
			subscription, err := cache.Provider.Subscribe("broadcast", "user_messages")
			if err != nil {
				log.Printf("Failed to subscribe: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for {
				msg, err := subscription.ReceiveMessage()
				if err != nil {
					log.Printf("Subscription error: %v", err)
					subscription.Close()
					break
				}

				switch msg.GetChannel() {
				case "broadcast":
					var broadcastMsg struct {
						Type      int    `json:"type"`
						Data      []byte `json:"data"`
						ChannelID string `json:"channel_id,omitempty"`
						Secure    bool   `json:"secure"`
					}

					json.Unmarshal([]byte(msg.GetPayload()), &broadcastMsg)

					if broadcastMsg.Secure && broadcastMsg.ChannelID != "" {
						var messageData map[string]interface{}
						json.Unmarshal(broadcastMsg.Data, &messageData)
						utils.BroadcastToChannel(broadcastMsg.ChannelID, messageData)
					} else {
						utils.BroadcastToAll(broadcastMsg.Type, broadcastMsg.Data)
					}

				case "user_messages":
					var userMsg struct {
						Type   int    `json:"type"`
						Data   []byte `json:"data"`
						UserID string `json:"user_id"`
					}

					json.Unmarshal([]byte(msg.GetPayload()), &userMsg)
					utils.SendToUser(userMsg.UserID, userMsg.Type, userMsg.Data)
				}
			}

			time.Sleep(5 * time.Second)
		}
	}()
}

func getCORSMiddleware() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     utils.AppConfig.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token", "X-Requested-With"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-CSRF-Token"},
		MaxAge:           3600,
	})
}

func main() {
	utils.Initialize()
	utils.CleanupUserStatusesOnServerStart()
	redisOptions := &redis.Options{
		Addr:         utils.AppConfig.RedisHost + ":" + utils.AppConfig.RedisPort,
		Password:     utils.AppConfig.RedisPassword,
		DB:           utils.AppConfig.RedisDB,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	err := cache.Initialize(redisOptions, utils.GetDB(), utils.Log)
	
	csrf.Initialize(cache.Provider, csrf.DefaultKeys)
	invite.Initialize(utils.GetDB())
	webhook.Initialize(utils.GetDB())

	startRedisSubscriber()
	utils.StartHeartbeat()
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	e := echo.New()

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = renderer

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(getCORSMiddleware())
	e.Use(middleware.Gzip())
	e.Use(utils.GetCSPMiddleware())

	e.Static("/public", "public")
	e.GET("/", func(c echo.Context) error {
		return c.File("templates/index.html")
	})
	e.GET("/login", func(c echo.Context) error {
		return c.Render(200, "login.html", nil)
	})
	e.GET("/register", func(c echo.Context) error {
		return c.Render(200, "register.html", nil)
	})

	publicroutes := []Route{
		{Method: "POST", Path: "/api/auth/register", Handler: api.RegisterHandler, AuthLevel: AUTH_NONE},
		{Method: "POST", Path: "/api/auth/login", Handler: api.LoginHandler, AuthLevel: AUTH_NONE},
		{Method: "POST", Path: "/api/auth/logout", Handler: api.LogoutHandler, AuthLevel: AUTH_NONE},
		{Method: "POST", Path: "/api/webhook/:webhookid/:token", Handler: api.ExecuteWebhookHandler, AuthLevel: AUTH_NONE},
	}

	wsRoutes := []Route{
		{Method: "GET", Path: "/api/ws", Handler: api.HandleAuthWebSocket, AuthLevel: AUTH_USER},
	}

	apiRoutes := []Route{
		{Method: "GET", Path: "/api/test", Handler: func(c echo.Context) error { return c.String(http.StatusOK, "hi") }, AuthLevel: AUTH_NONE},
		{Method: "GET", Path: "/api/user/getusers", Handler: api.GetUserList, AuthLevel: AUTH_ADMIN},
		{Method: "POST", Path: "/api/user/:username/make-admin", Handler: api.MakeUserAdmin, AuthLevel: AUTH_ADMIN},
		{Method: "POST", Path: "/api/user/:username/demote-admin", Handler: api.DemoteUserAdmin, AuthLevel: AUTH_ADMIN},
		{Method: "GET", Path: "/api/user/guilds", Handler: api.GetUserGuilds, AuthLevel: AUTH_USER},
		{Method: "POST", Path: "/api/ban/:userid", Handler: api.BanUserHandler, AuthLevel: AUTH_ADMIN},
		{Method: "POST", Path: "/api/unban/:userid", Handler: api.UnbanUserHandler, AuthLevel: AUTH_ADMIN},
		{Method: "GET", Path: "/api/csrf-token", Handler: api.GetCSRFToken, AuthLevel: AUTH_NONE},
		{Method: "POST", Path: "/api/guild/create", Handler: api.CreateGuildHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/guild/:id/members", Handler: api.GetGuildMembersHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/channels/get", Handler: api.GetChannelsHandler, AuthLevel: AUTH_USER},
		{Method: "POST", Path: "/api/channels/create", Handler: api.CreateChannelHandler, AuthLevel: AUTH_USER},
		{Method: "POST", Path: "/api/channels/delete", Handler: api.DeleteChannelHandler, AuthLevel: AUTH_USER},
		{Method: "DELETE", Path: "/api/webhook/delete/:webhookid", Handler: api.DeleteWebhookHandler, AuthLevel: AUTH_USER},
		{Method: "PUT", Path: "/api/channels/:channelid/edit", Handler: api.EditChannelHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/user/:id/profile", Handler: api.GetUserProfileHandler, AuthLevel: AUTH_NONE},
		{Method: "GET", Path: "/api/context/:type", Handler: api.GetContextMenuHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/channels/:channelid/messages", Handler: api.GetChannelMessagesHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/channels/:channelid/info", Handler: api.GetChannelInfoHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/guild/:id/ownership", Handler: api.CheckGuildOwnershipHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/guild/:guildId/permissions", Handler: api.GetGuildPermissionsHandler, AuthLevel: AUTH_USER},
		{Method: "POST", Path: "/api/webhook/create/:channelid", Handler: api.CreateWebhookHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/webhook/list/:channelid", Handler: api.ListWebhooksHandler, AuthLevel: AUTH_USER},
		{Method: "POST", Path: "/api/guild/leave/:id", Handler: api.LeaveGuildHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/embed", Handler: api.GetEmbedHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/v/:guildid/:channelid", Handler: api.ViewChannelHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/proxy-image", Handler: api.ProxyImageHandler, AuthLevel: AUTH_USER},
		{Method: "POST", Path: "/api/auth/refresh", Handler: api.RefreshSessionHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/v/:id", Handler: api.ViewGuildHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/v/main", Handler: api.ViewMainHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/guild/:guildid/info", Handler: api.GetGuildHandler, AuthLevel: AUTH_USER},
		{Method: "POST", Path: "/api/user/update-bio", Handler: api.UpdateBioHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/api/user/me", Handler: api.GetCurrentUser, AuthLevel: AUTH_USER},
		{Method: "POST", Path: "/api/user/update-username", Handler: api.UpdateUsernameHandler, AuthLevel: AUTH_USER},
		{Method: "GET", Path: "/i/:code", Handler: api.GetInviteInfoHandler, AuthLevel: AUTH_NONE},
		{Method: "POST", Path: "/api/invite/generate/:guild_id", Handler: api.GenerateInviteHandler, AuthLevel: AUTH_USER},
		{Method: "POST", Path: "/api/invite/join/:code", Handler: api.JoinByInviteHandler, AuthLevel: AUTH_USER},
		{Method: "POST", Path: "/api/user/profile-picture", Handler: api.UploadProfilePictureHandler, AuthLevel: AUTH_USER},
		{Method: "POST", Path: "/api/terminate/session/:sessionid", Handler: api.TerminateSessionHandler, AuthLevel: AUTH_ADMIN},
		{Method: "POST", Path: "/api/terminate/user/:userid", Handler: api.TerminateUserSessionsHandler, AuthLevel: AUTH_ADMIN},
		{Method: "GET", Path: "/admin", Handler: func(c echo.Context) error {
			return c.File("templates/admin.html")
		}, AuthLevel: AUTH_ADMIN},
	}

	RegisterRoutes(e, publicroutes, false)
	RegisterRoutes(e, apiRoutes, true)
	RegisterRoutes(e, wsRoutes, false)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}

		if code == http.StatusNotFound {
			c.Render(http.StatusNotFound, "404.html", nil)
			return
		}

		if code == http.StatusUnauthorized || code == http.StatusForbidden {
			c.Render(http.StatusForbidden, "403.html", nil)
			return
		}

		e.DefaultHTTPErrorHandler(err, c)
	}
	utils.StartHeartbeat()

	e.Logger.Fatal(e.Start(":8080"))
}
