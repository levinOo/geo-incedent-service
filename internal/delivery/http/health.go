package myHttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/levinOo/geo-incedent-service/internal/entity"
	"github.com/levinOo/geo-incedent-service/internal/service"
)

type HealthHandler interface {
	Check(c *gin.Context)
}

type HealthHandlerImpl struct {
	service *service.Service
}

func NewHealthHandler(service *service.Service) HealthHandler {
	return &HealthHandlerImpl{service: service}
}

// Check godoc
// @Summary Проверяет здоровье сервиса
// @Description Метод для проверки здоровья сервиса
// @Tags health
// @Produce json
// @Success 200 {object} entity.HealthResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /health [get]
func (h *HealthHandlerImpl) Check(c *gin.Context) {
	resp, err := h.service.Health.Check(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error:   "Не удалось проверить здоровье сервиса",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
