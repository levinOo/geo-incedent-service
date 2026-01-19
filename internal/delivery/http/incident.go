package myHttp

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/levinOo/geo-incedent-service/internal/entity"
	"github.com/levinOo/geo-incedent-service/internal/service"
)

type IncidentHandler interface {
	CreateIncident(c *gin.Context)
	GetIncidents(c *gin.Context)
	GetIncident(c *gin.Context)
	UpdateIncident(c *gin.Context)
	DeleteIncident(c *gin.Context)
	GetStats(c *gin.Context)
}

type IncidentHandlerImpl struct {
	service *service.Service
}

func NewIncidentHandler(service *service.Service) IncidentHandler {
	return &IncidentHandlerImpl{service: service}
}

// CreateIncident godoc
// @Summary Создает новый инцидент
// @Description Метод для создания инцидента. Создает инцидент с названием, описанием и гео-зоной (Polygon). Зона определяет опасную область для проверок локаций.
// @Tags incidents
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param incident body entity.CreateIncidentRequest true "Incident data"
// @Success 201 {object} entity.IncidentResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /incidents [post]
func (h *IncidentHandlerImpl) CreateIncident(c *gin.Context) {
	var req entity.CreateIncidentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "Некорректное тело запроса",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.service.Incident.Create(c, &req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error:   "Не удалось создать инцидент",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetIncident godoc
// @Summary Получает инцидент по ID
// @Description Метод для получения инцидента по уникальному идентификатору (UUID). ID передается в URL как параметр пути.
// @Tags incidents
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Incident ID"
// @Success 200 {object} entity.GetIncidentResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /incidents/{id} [get]
func (h *IncidentHandlerImpl) GetIncident(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.service.Incident.FindByID(c, id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error:   "Не удалось получить инцидент",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetIncidents godoc
// @Summary Получает список активных инцидентов
// @Description Метод для получения пагенированного списка инцидентов. Поддерживает параметры limit и offset.
// @Tags incidents
// @Produce json
// @Security ApiKeyAuth
// @Param limit query int false "Количество записей"
// @Param offset query int false "Смещение (для пагинации)"
// @Success 200 {object} entity.GetIncidentsResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /incidents [get]
func (h *IncidentHandlerImpl) GetIncidents(c *gin.Context) {
	limit := c.Query("limit")
	offset := c.Query("offset")

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 0 {
		limitInt = 10
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		offsetInt = 0
	}

	resp, err := h.service.Incident.FindAll(c, limitInt, offsetInt)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error:   "Не удалось получить список инцидентов",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateIncident godoc
// @Summary Обновляет инцидент
// @Description Метод для обновления инцидента по уникальному идентификатору (UUID). ID передается в URL как параметр пути. Можно обновить только название, описание и гео-зону (Polygon).
// @Tags incidents
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Incident ID"
// @Param incident body entity.UpdateIncidentRequest true "Incident"
// @Success 200 {object} entity.IncidentResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /incidents/{id} [put]
func (h *IncidentHandlerImpl) UpdateIncident(c *gin.Context) {
	id := c.Param("id")
	var req entity.UpdateIncidentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "Некорректное тело запроса",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.service.Incident.Update(c, &req, id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error:   "Не удалось обновить инцидент",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteIncident godoc
// @Summary Деактивирует инцидент
// @Description Метод для деактивации инцидента по уникальному идентификатору (UUID). ID передается в URL как параметр пути.
// @Tags incidents
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Incident ID"
// @Success 200 {object} entity.IncidentResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /incidents/{id} [delete]
func (h *IncidentHandlerImpl) DeleteIncident(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.service.Incident.Delete(c, id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error:   "Не удалось удалить инцидент",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetStats godoc
// @Summary Получает статистику инцидентов
// @Description Получает статистику инцидентов
// @Tags incidents
// @Produce json
// @Success 200 {object} entity.StatsResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /incidents/stats [get]
func (h *IncidentHandlerImpl) GetStats(c *gin.Context) {
	stats, err := h.service.Incident.GetStats(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Error:   "Не удалось получить статистику",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
