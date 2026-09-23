package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

type ImportListResponse struct {
	Data []domain.Import `json:"data"`
	Meta PaginationMeta  `json:"meta"`
}

type ImportResponse struct {
	Data domain.Import `json:"data"`
}

// createImport enregistre un import avec ses lignes brutes.
//
//	@Summary		Créer un import
//	@Description	Enregistre un lot de lignes pour un type de ressource au statut draft.
//	@Tags			Administration
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.ImportCreate	true	"Import à créer"
//	@Success		201		{object}	ImportResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Failure		403		{object}	Problem
//	@Failure		500		{object}	Problem
//	@Router			/api/v1/admin/imports [post]
func (h *Handler) createImport(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		return
	}
	var input domain.ImportCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	created, err := h.service.CreateImport(c.Request.Context(), userID, input)
	if err != nil {
		handleServiceError(c, err, "Impossible de créer l'import.")
		return
	}
	c.JSON(http.StatusCreated, ImportResponse{Data: created})
}

// listImports liste les imports, du plus récent au plus ancien.
//
//	@Summary		Lister les imports
//	@Tags			Administration
//	@Security		BearerAuth
//	@Produce		json
//	@Param			limit	query	integer	false	"Nombre de résultats" default(20) minimum(1) maximum(100)
//	@Param			offset	query	integer	false	"Décalage" default(0) minimum(0)
//	@Success		200		{object}	ImportListResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Failure		403		{object}	Problem
//	@Failure		500		{object}	Problem
//	@Router			/api/v1/admin/imports [get]
func (h *Handler) listImports(c *gin.Context) {
	limit, offset, ok := pagination(c)
	if !ok {
		return
	}
	imports, err := h.service.ListImports(c.Request.Context(), domain.Pagination{Limit: limit, Offset: offset})
	if err != nil {
		handleServiceError(c, err, "Impossible de lister les imports.")
		return
	}
	c.JSON(http.StatusOK, ImportListResponse{Data: imports, Meta: PaginationMeta{Limit: limit, Offset: offset}})
}

// getImport retourne l'état et le rapport d'erreurs d'un import.
//
//	@Summary		Lire un import
//	@Description	Retourne l'état (draft, validated, published) et le rapport d'erreurs.
//	@Tags			Administration
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"UUID de l'import" format(uuid)
//	@Success		200	{object}	ImportResponse
//	@Failure		400	{object}	Problem
//	@Failure		401	{object}	Problem
//	@Failure		403	{object}	Problem
//	@Failure		404	{object}	Problem
//	@Failure		500	{object}	Problem
//	@Router			/api/v1/admin/imports/{id} [get]
func (h *Handler) getImport(c *gin.Context) {
	imported, err := h.service.GetImport(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "Impossible de charger l'import.")
		return
	}
	c.JSON(http.StatusOK, ImportResponse{Data: imported})
}

// validateImport lance la validation de toutes les lignes d'un import.
//
//	@Summary		Valider un import
//	@Description	Valide chaque ligne et mémorise le rapport d'erreurs. Réservé aux imports au statut draft.
//	@Tags			Administration
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path	string	true	"UUID de l'import" format(uuid)
//	@Success		200	{object}	ImportResponse
//	@Failure		400	{object}	Problem
//	@Failure		401	{object}	Problem
//	@Failure		403	{object}	Problem
//	@Failure		404	{object}	Problem
//	@Failure		409	{object}	Problem
//	@Failure		500	{object}	Problem
//	@Router			/api/v1/admin/imports/{id}/validate [post]
func (h *Handler) validateImport(c *gin.Context) {
	imported, err := h.service.ValidateImport(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "Impossible de valider l'import.")
		return
	}
	c.JSON(http.StatusOK, ImportResponse{Data: imported})
}

// publishImport applique les lignes valides via les upserts métier.
//
//	@Summary		Publier un import
//	@Description	Applique les lignes valides. Réservé aux imports au statut validated.
//	@Tags			Administration
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path	string	true	"UUID de l'import" format(uuid)
//	@Success		200	{object}	ImportResponse
//	@Failure		400	{object}	Problem
//	@Failure		401	{object}	Problem
//	@Failure		403	{object}	Problem
//	@Failure		404	{object}	Problem
//	@Failure		409	{object}	Problem
//	@Failure		500	{object}	Problem
//	@Router			/api/v1/admin/imports/{id}/publish [post]
func (h *Handler) publishImport(c *gin.Context) {
	imported, err := h.service.PublishImport(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "Impossible de publier l'import.")
		return
	}
	c.JSON(http.StatusOK, ImportResponse{Data: imported})
}
