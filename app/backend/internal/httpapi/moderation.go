package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

type ModerationQueueResponse struct {
	Data []domain.ModerationQueueItem `json:"data"`
	Meta PaginationMeta               `json:"meta"`
}

type ModerationDecisionBody struct {
	Note       *string `json:"note,omitempty" example:"Prix cohérent avec les relevés voisins"`
	ReasonCode *string `json:"reason_code,omitempty" example:"outlier"`
}

// listModerationQueue liste les contributions à contrôler, par statut (pending par défaut).
//
//	@Summary		Lister la file de modération
//	@Description	Contributeur, produit ou carburant, montant, lieu et état de chaque contribution à contrôler.
//	@Tags			Modération
//	@Security		BearerAuth
//	@Produce		json
//	@Param			status	query		string	false	"Statut (pending, approved, rejected, flagged)"
//	@Param			limit	query		integer	false	"Nombre de résultats" default(20) minimum(1) maximum(100)
//	@Param			offset	query		integer	false	"Décalage" default(0) minimum(0)
//	@Success		200		{object}	ModerationQueueResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Failure		403		{object}	Problem
//	@Router			/api/v1/moderation/queue [get]
func (h *Handler) listModerationQueue(c *gin.Context) {
	limit, offset, ok := pagination(c)
	if !ok {
		return
	}
	items, err := h.service.ListModerationQueue(c.Request.Context(), domain.ModerationFilter{
		Status: c.Query("status"),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		handleServiceError(c, err, "Impossible de lister la file de modération.")
		return
	}
	c.JSON(http.StatusOK, ModerationQueueResponse{Data: items, Meta: PaginationMeta{Limit: limit, Offset: offset}})
}

// approveContribution approuve une contribution en attente.
//
//	@Summary		Approuver une contribution
//	@Tags			Modération
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string					true	"UUID de la contribution"	format(uuid)
//	@Param			body	body	ModerationDecisionBody	false	"Note facultative"
//	@Success		200		{object}	ContributionDetailResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Failure		403		{object}	Problem
//	@Failure		404		{object}	Problem
//	@Failure		409		{object}	Problem
//	@Router			/api/v1/moderation/contributions/{id}/approve [post]
func (h *Handler) approveContribution(c *gin.Context) {
	h.reviewContribution(c, domain.StatusApproved)
}

// rejectContribution rejette une contribution en attente avec un motif obligatoire.
//
//	@Summary		Rejeter une contribution
//	@Description	Le rejet exige une note expliquant le motif.
//	@Tags			Modération
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string					true	"UUID de la contribution"	format(uuid)
//	@Param			body	body	ModerationDecisionBody	false	"Motif du rejet"
//	@Success		200		{object}	ContributionDetailResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Failure		403		{object}	Problem
//	@Failure		404		{object}	Problem
//	@Failure		409		{object}	Problem
//	@Router			/api/v1/moderation/contributions/{id}/reject [post]
func (h *Handler) rejectContribution(c *gin.Context) {
	h.reviewContribution(c, domain.StatusRejected)
}

func (h *Handler) reviewContribution(c *gin.Context, decision string) {
	moderatorID, ok := currentUserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		return
	}
	var body ModerationDecisionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	var note *string
	if body.Note != nil {
		trimmed := strings.TrimSpace(*body.Note)
		note = &trimmed
	}
	detail, err := h.service.ReviewContribution(c.Request.Context(), moderatorID, c.Param("id"), domain.ContributionReviewInput{
		Decision:   decision,
		Note:       note,
		ReasonCode: body.ReasonCode,
	})
	if err != nil {
		handleServiceError(c, err, "Impossible d'appliquer la décision.")
		return
	}
	c.JSON(http.StatusOK, ContributionDetailResponse{Data: detail})
}
