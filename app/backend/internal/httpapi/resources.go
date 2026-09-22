package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
	"github.com/laurentpoirierfr/ojozone/internal/service"
)

type ResourceListResponse struct {
	Data any            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}
type ResourceResponse struct {
	Data any `json:"data"`
}

func registerResourceRoutes(api *gin.RouterGroup, h *Handler) {
	resources := []struct {
		name    string
		handler gin.HandlerFunc
	}{
		{"geo-areas", h.geoAreas}, {"sources", h.sources}, {"categories", h.categories}, {"merchants", h.merchants}, {"locations", h.locations},
		{"fuel-types", h.fuelTypes}, {"fuel-prices", h.fuelPrices}, {"housing-observations", h.housingObservations},
		{"income-observations", h.incomeObservations}, {"evidence-files", h.evidenceFiles},
	}
	for _, resource := range resources {
		api.GET("/"+resource.name, resource.handler)
		api.POST("/"+resource.name, h.authenticate, requireRole(domain.RoleAdmin), resource.handler)
		api.GET("/"+resource.name+"/:id", resource.handler)
		api.PUT("/"+resource.name+"/:id", h.authenticate, requireRole(domain.RoleAdmin), resource.handler)
		api.DELETE("/"+resource.name+"/:id", h.authenticate, requireRole(domain.RoleAdmin), resource.handler)
	}
	api.GET("/units", h.units)
	api.POST("/units", h.authenticate, requireRole(domain.RoleAdmin), h.units)
	api.GET("/units/:code", h.units)
	api.PUT("/units/:code", h.authenticate, requireRole(domain.RoleAdmin), h.units)
	api.DELETE("/units/:code", h.authenticate, requireRole(domain.RoleAdmin), h.units)
	aggregatePath := "/price-aggregates/:geo_area_id/:metric_type/:subject_id/:month"
	api.GET("/price-aggregates", h.priceAggregates)
	api.POST("/price-aggregates", h.authenticate, requireRole(domain.RoleAdmin), h.priceAggregates)
	api.GET(aggregatePath, h.priceAggregates)
	api.PUT(aggregatePath, h.authenticate, requireRole(domain.RoleAdmin), h.priceAggregates)
	api.DELETE(aggregatePath, h.authenticate, requireRole(domain.RoleAdmin), h.priceAggregates)
	api.GET("/moderation-events", h.moderationEvents)
	api.POST("/moderation-events", h.authenticate, requireRole(domain.RoleModerator, domain.RoleAdmin), h.moderationEvents)
	api.GET("/moderation-events/:id", h.moderationEvents)
	api.GET("/admin/users", h.authenticate, requireRole(domain.RoleAdmin), h.adminUsers)
	api.GET("/admin/users/:id", h.authenticate, requireRole(domain.RoleAdmin), h.adminUsers)
}

func (h *Handler) handleResource(c *gin.Context, resource string) {
	key := domain.ResourceKey{ID: c.Param("id"), Code: c.Param("code"), GeoAreaID: c.Param("geo_area_id"), MetricType: c.Param("metric_type"), SubjectID: c.Param("subject_id"), Month: c.Param("month")}
	isItem := key.ID != "" || key.Code != "" || key.GeoAreaID != ""
	switch c.Request.Method {
	case http.MethodGet:
		if isItem {
			value, err := h.service.GetResource(c.Request.Context(), resource, key)
			if err != nil {
				handleServiceError(c, err, "Impossible de charger la ressource.")
				return
			}
			c.JSON(http.StatusOK, ResourceResponse{Data: value})
			return
		}
		limit, offset, ok := pagination(c)
		if !ok {
			return
		}
		value, err := h.service.ListResource(c.Request.Context(), resource, domain.Pagination{Limit: limit, Offset: offset})
		if err != nil {
			handleServiceError(c, err, "Impossible de lister les ressources.")
			return
		}
		c.JSON(http.StatusOK, ResourceListResponse{Data: value, Meta: PaginationMeta{Limit: limit, Offset: offset}})
	case http.MethodPost, http.MethodPut:
		raw, err := c.GetRawData()
		if err != nil || !json.Valid(raw) {
			writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
			return
		}
		var value any
		if c.Request.Method == http.MethodPost {
			value, err = h.service.UpsertResource(c.Request.Context(), resource, raw)
		} else {
			value, err = h.service.ReplaceResource(c.Request.Context(), resource, key, raw)
		}
		if err != nil {
			handleServiceError(c, err, "Impossible d'enregistrer la ressource.")
			return
		}
		c.JSON(http.StatusOK, ResourceResponse{Data: value})
	case http.MethodDelete:
		if err := h.service.DeleteResource(c.Request.Context(), resource, key); err != nil {
			handleServiceError(c, err, "Impossible de supprimer la ressource.")
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// @Summary Gérer les zones géographiques
// @Tags Zones géographiques
// @Accept json
// @Produce json
// @Param payload body domain.GeoAreaUpsert false "Zone géographique"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/geo-areas [get]
// @Router /api/v1/geo-areas [post]
// @Router /api/v1/geo-areas/{id} [get]
// @Router /api/v1/geo-areas/{id} [put]
// @Router /api/v1/geo-areas/{id} [delete]
func (h *Handler) geoAreas(c *gin.Context) { h.handleResource(c, service.ResourceGeoAreas) }

// @Summary Gérer les sources
// @Tags Sources
// @Accept json
// @Produce json
// @Param payload body domain.SourceUpsert false "Source"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/sources [get]
// @Router /api/v1/sources [post]
// @Router /api/v1/sources/{id} [get]
// @Router /api/v1/sources/{id} [put]
// @Router /api/v1/sources/{id} [delete]
func (h *Handler) sources(c *gin.Context) { h.handleResource(c, service.ResourceSources) }

// @Summary Gérer les catégories
// @Tags Catégories
// @Accept json
// @Produce json
// @Param payload body domain.CategoryUpsert false "Catégorie"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/categories [get]
// @Router /api/v1/categories [post]
// @Router /api/v1/categories/{id} [get]
// @Router /api/v1/categories/{id} [put]
// @Router /api/v1/categories/{id} [delete]
func (h *Handler) categories(c *gin.Context) { h.handleResource(c, service.ResourceCategories) }

// @Summary Gérer les unités
// @Tags Unités
// @Accept json
// @Produce json
// @Param payload body domain.UnitUpsert false "Unité"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/units [get]
// @Router /api/v1/units [post]
// @Router /api/v1/units/{code} [get]
// @Router /api/v1/units/{code} [put]
// @Router /api/v1/units/{code} [delete]
func (h *Handler) units(c *gin.Context) { h.handleResource(c, service.ResourceUnits) }

// @Summary Gérer les commerçants
// @Tags Commerçants
// @Accept json
// @Produce json
// @Param payload body domain.MerchantUpsert false "Commerçant"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/merchants [get]
// @Router /api/v1/merchants [post]
// @Router /api/v1/merchants/{id} [get]
// @Router /api/v1/merchants/{id} [put]
// @Router /api/v1/merchants/{id} [delete]
func (h *Handler) merchants(c *gin.Context) { h.handleResource(c, service.ResourceMerchants) }

// @Summary Gérer les lieux
// @Tags Lieux
// @Accept json
// @Produce json
// @Param payload body domain.LocationUpsert false "Lieu avec latitude et longitude"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/locations [get]
// @Router /api/v1/locations [post]
// @Router /api/v1/locations/{id} [get]
// @Router /api/v1/locations/{id} [put]
// @Router /api/v1/locations/{id} [delete]
func (h *Handler) locations(c *gin.Context) { h.handleResource(c, service.ResourceLocations) }

// @Summary Gérer les types de carburant
// @Tags Carburants
// @Accept json
// @Produce json
// @Param payload body domain.FuelTypeUpsert false "Type de carburant"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/fuel-types [get]
// @Router /api/v1/fuel-types [post]
// @Router /api/v1/fuel-types/{id} [get]
// @Router /api/v1/fuel-types/{id} [put]
// @Router /api/v1/fuel-types/{id} [delete]
func (h *Handler) fuelTypes(c *gin.Context) { h.handleResource(c, service.ResourceFuelTypes) }

// @Summary Gérer les prix des carburants
// @Tags Prix carburants
// @Accept json
// @Produce json
// @Param payload body domain.FuelPriceUpsert false "Observation"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/fuel-prices [get]
// @Router /api/v1/fuel-prices [post]
// @Router /api/v1/fuel-prices/{id} [get]
// @Router /api/v1/fuel-prices/{id} [put]
// @Router /api/v1/fuel-prices/{id} [delete]
func (h *Handler) fuelPrices(c *gin.Context) { h.handleResource(c, service.ResourceFuelPrices) }

// @Summary Gérer les observations immobilières
// @Tags Immobilier
// @Accept json
// @Produce json
// @Param payload body domain.HousingUpsert false "Observation"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/housing-observations [get]
// @Router /api/v1/housing-observations [post]
// @Router /api/v1/housing-observations/{id} [get]
// @Router /api/v1/housing-observations/{id} [put]
// @Router /api/v1/housing-observations/{id} [delete]
func (h *Handler) housingObservations(c *gin.Context) { h.handleResource(c, service.ResourceHousing) }

// @Summary Gérer les observations de revenus
// @Tags Revenus
// @Accept json
// @Produce json
// @Param payload body domain.IncomeUpsert false "Observation"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/income-observations [get]
// @Router /api/v1/income-observations [post]
// @Router /api/v1/income-observations/{id} [get]
// @Router /api/v1/income-observations/{id} [put]
// @Router /api/v1/income-observations/{id} [delete]
func (h *Handler) incomeObservations(c *gin.Context) { h.handleResource(c, service.ResourceIncome) }

// @Summary Gérer les preuves
// @Tags Preuves
// @Accept json
// @Produce json
// @Param payload body domain.EvidenceUpsert false "Preuve"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/evidence-files [get]
// @Router /api/v1/evidence-files [post]
// @Router /api/v1/evidence-files/{id} [get]
// @Router /api/v1/evidence-files/{id} [put]
// @Router /api/v1/evidence-files/{id} [delete]
func (h *Handler) evidenceFiles(c *gin.Context) { h.handleResource(c, service.ResourceEvidence) }

// @Summary Gérer les agrégats mensuels
// @Tags Agrégats
// @Accept json
// @Produce json
// @Param payload body domain.PriceAggregateUpsert false "Agrégat"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 409 {object} Problem
// @Router /api/v1/price-aggregates [get]
// @Router /api/v1/price-aggregates [post]
// @Router /api/v1/price-aggregates/{geo_area_id}/{metric_type}/{subject_id}/{month} [get]
// @Router /api/v1/price-aggregates/{geo_area_id}/{metric_type}/{subject_id}/{month} [put]
// @Router /api/v1/price-aggregates/{geo_area_id}/{metric_type}/{subject_id}/{month} [delete]
func (h *Handler) priceAggregates(c *gin.Context) { h.handleResource(c, service.ResourceAggregates) }

// @Summary Consulter ou ajouter les événements de modération
// @Description Ressource append-only : aucun PUT ni DELETE.
// @Tags Modération
// @Accept json
// @Produce json
// @Param payload body domain.ModerationEventAppend false "Événement"
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/moderation-events [get]
// @Router /api/v1/moderation-events [post]
// @Router /api/v1/moderation-events/{id} [get]
func (h *Handler) moderationEvents(c *gin.Context) { h.handleResource(c, service.ResourceModeration) }

// @Summary Consulter les utilisateurs en administration
// @Description Lecture seule en l'absence d'authentification ; password_hash n'est jamais exposé.
// @Tags Administration
// @Produce json
// @Success 200 {object} ResourceResponse
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/admin/users [get]
// @Router /api/v1/admin/users/{id} [get]
func (h *Handler) adminUsers(c *gin.Context) { h.handleResource(c, service.ResourceAdminUsers) }
