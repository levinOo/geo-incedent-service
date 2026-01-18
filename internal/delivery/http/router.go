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

	h := NewHandler(service)

	// Swagger UI
	r.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := r.Group("/api/v1")
	{
		r.Use(LoggingMiddleware())

		health := api.Group("system")
		{
			health.GET("/health", h.Health.Check)
		}

		incidents := api.Group("incidents")
		{
			r.Use(ApiKeyMiddleware(cfg))

			incidents.POST("", h.Incident.CreateIncident)
			incidents.GET("", h.Incident.GetIncidents)
			incidents.GET("/:id", h.Incident.GetIncident)
			incidents.PATCH("/:id", h.Incident.UpdateIncident)
			incidents.DELETE("/:id", h.Incident.DeleteIncident)

			stats := api.Group("stats")
			{
				stats.GET("", h.Incident.GetStats)
			}
		}

		location := api.Group("location")
		{
			location.POST("", h.Location.CheckLocation)
		}
	}

	return r
}
