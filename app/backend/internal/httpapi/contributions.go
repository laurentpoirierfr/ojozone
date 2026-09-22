package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/laurentpoirierfr/ojozone/internal/auth"
	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

type ContributionResultResponse struct {
	Data domain.ContributionResult `json:"data"`
}

type ContributionDetailResponse struct {
	Data domain.ContributionDetail `json:"data"`
}

// submitProductPriceContribution propose un prix produit avec le statut pending.
//
//	@Summary		Proposer un prix produit
//	@Description	La soumission d'un membre crée une observation au statut pending, à modérer.
//	@Tags			Contributions
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.ProductContributionInput	true	"Prix observé"
//	@Success		202		{object}	ContributionResultResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Failure		403		{object}	Problem
//	@Failure		500		{object}	Problem
//	@Router			/api/v1/contributions/product-prices [post]
func (h *Handler) submitProductPriceContribution(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		return
	}
	var input domain.ProductContributionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	result, err := h.service.SubmitProductContribution(c.Request.Context(), userID, input)
	if err != nil {
		handleServiceError(c, err, "Impossible de soumettre la contribution.")
		return
	}
	c.JSON(http.StatusAccepted, ContributionResultResponse{Data: result})
}

// submitFuelPriceContribution propose un prix carburant avec le statut pending.
//
//	@Summary		Proposer un prix carburant
//	@Description	La soumission d'un membre crée une observation au statut pending, à modérer.
//	@Tags			Contributions
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.FuelContributionInput	true	"Prix au litre observé"
//	@Success		202		{object}	ContributionResultResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Failure		403		{object}	Problem
//	@Failure		500		{object}	Problem
//	@Router			/api/v1/contributions/fuel-prices [post]
func (h *Handler) submitFuelPriceContribution(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		return
	}
	var input domain.FuelContributionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	result, err := h.service.SubmitFuelContribution(c.Request.Context(), userID, input)
	if err != nil {
		handleServiceError(c, err, "Impossible de soumettre la contribution.")
		return
	}
	c.JSON(http.StatusAccepted, ContributionResultResponse{Data: result})
}

// getContribution expose une contribution à son propriétaire ou à la modération.
//
//	@Summary		Consulter le statut d'une contribution
//	@Tags			Contributions
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"UUID de la contribution"	format(uuid)
//	@Success		200	{object}	ContributionDetailResponse
//	@Failure		401	{object}	Problem
//	@Failure		403	{object}	Problem
//	@Failure		404	{object}	Problem
//	@Router			/api/v1/contributions/{id} [get]
func (h *Handler) getContribution(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		return
	}
	role := claimsRole(c)
	detail, err := h.service.GetContribution(c.Request.Context(), userID, role, c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "Impossible de charger la contribution.")
		return
	}
	c.JSON(http.StatusOK, ContributionDetailResponse{Data: detail})
}

// patchContribution corrige une contribution pending appartenant à l'utilisateur courant.
//
//	@Summary		Corriger une contribution
//	@Description	La contribution doit encore être au statut pending et appartenir à l'utilisateur courant.
//	@Tags			Contributions
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"UUID de la contribution"	format(uuid)
//	@Param			input	body		domain.ContributionPatch	true	"Champs à corriger"
//	@Success		200		{object}	ContributionDetailResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Failure		403		{object}	Problem
//	@Failure		404		{object}	Problem
//	@Failure		409		{object}	Problem
//	@Router			/api/v1/contributions/{id} [patch]
func (h *Handler) patchContribution(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		return
	}
	var input domain.ContributionPatch
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	detail, err := h.service.UpdateContribution(c.Request.Context(), userID, c.Param("id"), input)
	if err != nil {
		handleServiceError(c, err, "Impossible de corriger la contribution.")
		return
	}
	c.JSON(http.StatusOK, ContributionDetailResponse{Data: detail})
}

// deleteContribution retire une contribution pending appartenant à l'utilisateur courant.
//
//	@Summary		Retirer une contribution
//	@Description	La contribution doit encore être au statut pending et appartenir à l'utilisateur courant.
//	@Tags			Contributions
//	@Security		BearerAuth
//	@Param			id	path	string	true	"UUID de la contribution"	format(uuid)
//	@Success		204
//	@Failure		401	{object}	Problem
//	@Failure		403	{object}	Problem
//	@Failure		404	{object}	Problem
//	@Failure		409	{object}	Problem
//	@Router			/api/v1/contributions/{id} [delete]
func (h *Handler) deleteContribution(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		return
	}
	if err := h.service.DeleteContribution(c.Request.Context(), userID, c.Param("id")); err != nil {
		handleServiceError(c, err, "Impossible de retirer la contribution.")
		return
	}
	c.Status(http.StatusNoContent)
}

func claimsRole(c *gin.Context) string {
	value, exists := c.Get(claimsKey)
	if !exists {
		return ""
	}
	claims, ok := value.(auth.Claims)
	if !ok {
		return ""
	}
	return claims.Role
}
