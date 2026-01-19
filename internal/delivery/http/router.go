package myHttp

import (
	"github.com/gin-gonic/gin"
	"github.com/levinOo/geo-incedent-service/config"
	_ "github.com/levinOo/geo-incedent-service/docs"
	"github.com/levinOo/geo-incedent-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(cfg *config.HTTPServerConfig, service *service.Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(LoggingMiddleware())

	h := NewHandler(service)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		api.GET("/system/health", ApiKeyMiddleware(cfg), h.Health.Check)

		incidents := api.Group("/incidents")
		incidents.Use(ApiKeyMiddleware(cfg))
		{
			incidents.GET("/stats", h.Incident.GetStats)
			incidents.POST("", h.Incident.CreateIncident)
			incidents.GET("", h.Incident.GetIncidents)
			incidents.GET("/:id", h.Incident.GetIncident)
			incidents.PUT("/:id", h.Incident.UpdateIncident)
			incidents.DELETE("/:id", h.Incident.DeleteIncident)
		}

		location := api.Group("/location")
		{
			location.POST("/check", h.Location.CheckLocation)
		}
	}

	return r
}
