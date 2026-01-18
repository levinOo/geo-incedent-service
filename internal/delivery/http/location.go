package myHttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/levinOo/geo-incedent-service/internal/entity"
	"github.com/levinOo/geo-incedent-service/internal/service"
)

type LocationHandler interface {
	CheckLocation(c *gin.Context)
}

type LocationHandlerImpl struct {
	service *service.Service
}

func NewLocationHandler(service *service.Service) LocationHandler {
	return &LocationHandlerImpl{service: service}
}

// CheckLocation godoc
// @Summary Проверяет локацию
// @Description Метод проверяет находится ли пользователь в опасной зоне. Принимает координаты пользователя и userID.
// @Tags location
// @Accept json
// @Produce json
// @Param location body entity.CheckLocationRequest true "User data"
// @Success 200 {object} entity.CheckLocationResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /location [post]
func (h *LocationHandlerImpl) CheckLocation(c *gin.Context) {
	var req *entity.CheckLocationRequest

	if err := c.ShouldBindJSON(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "Некорректное тело запроса",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.service.Location.CheckLocation(c, req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error:   "Не удалось проверить локацию",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)

}
