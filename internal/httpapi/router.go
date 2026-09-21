package httpapi

import (
	"log/slog"

	"github.com/KouroshEsmaeili/go-rest-api-template/internal/health"
	"github.com/KouroshEsmaeili/go-rest-api-template/internal/posts"
	"github.com/gin-gonic/gin"
)

func NewRouter(environment string, store posts.Store, pinger health.Pinger, logger *slog.Logger) *gin.Engine {
	switch environment {
	case "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Error("failed to configure trusted proxies", "error", err)
	}

	healthHandler := health.NewHandler(pinger)
	postHandler := NewPostHandler(store, logger)

	router.GET("/health", healthHandler.Live)
	router.GET("/ready", healthHandler.Ready)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/posts", postHandler.List)
		v1.GET("/posts/:id", postHandler.Get)
		v1.POST("/posts", postHandler.Create)
		v1.PUT("/posts/:id", postHandler.Update)
		v1.DELETE("/posts/:id", postHandler.Delete)
	}

	router.NoRoute(func(c *gin.Context) {
		writeError(c, 404, "not_found", "route not found")
	})
	return router
}
