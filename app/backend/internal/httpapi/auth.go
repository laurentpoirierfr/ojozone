package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/laurentpoirierfr/ojozone/internal/auth"
	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

const (
	// claimsKey stocke les revendications du jeton d'accès dans le contexte Gin.
	claimsKey = "auth.claims"
	// bearerPrefix est le schéma attendu dans Authorization.
	bearerPrefix = "Bearer "
)

// tokenParser vérifie un jeton d'accès. *auth.Manager l'implémente.
type tokenParser interface {
	ParseAccessToken(string) (auth.Claims, error)
}

type UserResponse struct {
	Data domain.User `json:"data"`
}

type AuthResponse struct {
	Data domain.AuthResult `json:"data"`
}

type ContributionsResponse struct {
	Data []domain.Contribution `json:"data"`
	Meta PaginationMeta        `json:"meta"`
}

// authenticate exige un jeton d'accès valide puis place ses revendications dans le contexte.
func (h *Handler) authenticate(c *gin.Context) {
	header := strings.TrimSpace(c.GetHeader("Authorization"))
	token, ok := strings.CutPrefix(header, bearerPrefix)
	if !ok || strings.TrimSpace(token) == "" {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		c.Abort()
		return
	}
	claims, err := h.tokens.ParseAccessToken(strings.TrimSpace(token))
	if err != nil {
		writeProblem(c, http.StatusUnauthorized, "invalid_token", "Jeton d'accès invalide ou expiré.")
		c.Abort()
		return
	}
	c.Set(claimsKey, claims)
	c.Next()
}

// requireRole restreint la route aux rôles autorisés après authentication.
func requireRole(allowed ...string) gin.HandlerFunc {
	roles := make(map[string]bool, len(allowed))
	for _, role := range allowed {
		roles[role] = true
	}
	return func(c *gin.Context) {
		value, exists := c.Get(claimsKey)
		if !exists {
			writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
			c.Abort()
			return
		}
		claims, ok := value.(auth.Claims)
		if !ok || !roles[claims.Role] {
			writeProblem(c, http.StatusForbidden, "forbidden", "Vos droits ne permettent pas cette opération.")
			c.Abort()
			return
		}
		c.Next()
	}
}

func currentUserID(c *gin.Context) (string, bool) {
	value, exists := c.Get(claimsKey)
	if !exists {
		return "", false
	}
	claims, ok := value.(auth.Claims)
	if !ok {
		return "", false
	}
	return claims.Subject, true
}

// register crée un compte membre et ouvre une session.
//
//	@Summary		Créer un compte et ouvrir une session
//	@Description	Valide l'email, le mot de passe et la langue, puis retourne le profil et deux jetons.
//	@Tags			Authentification
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.RegisterInput	true	"Inscription"
//	@Success		201		{object}	AuthResponse
//	@Failure		400		{object}	Problem
//	@Failure		409		{object}	Problem
//	@Failure		500		{object}	Problem
//	@Router			/api/v1/auth/register [post]
func (h *Handler) register(c *gin.Context) {
	var input domain.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	result, err := h.service.Register(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, err, "Impossible de créer le compte.")
		return
	}
	c.JSON(http.StatusCreated, AuthResponse{Data: result})
}

// login ouvre une session avec email et mot de passe.
//
//	@Summary		Ouvrir une session
//	@Tags			Authentification
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.LoginInput	true	"Identifiants"
//	@Success		200		{object}	AuthResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Router			/api/v1/auth/login [post]
func (h *Handler) login(c *gin.Context) {
	var input domain.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	result, err := h.service.Login(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, err, "Impossible de se connecter.")
		return
	}
	c.JSON(http.StatusOK, AuthResponse{Data: result})
}

// refresh renouvelle les jetons par rotation.
//
//	@Summary		Renouveler les jetons
//	@Description	Un jeton de rafraîchissement ne peut être utilisé qu'une seule fois.
//	@Tags			Authentification
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.RefreshInput	true	"Jeton de rafraîchissement"
//	@Success		200		{object}	AuthResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Router			/api/v1/auth/refresh [post]
func (h *Handler) refresh(c *gin.Context) {
	var input domain.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	result, err := h.service.Refresh(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, err, "Impossible de renouveler les jetons.")
		return
	}
	c.JSON(http.StatusOK, AuthResponse{Data: result})
}

// logout révoque la session associée au jeton de rafraîchissement.
//
//	@Summary		Révoquer la session
//	@Tags			Authentification
//	@Accept			json
//	@Param			input	body	domain.RefreshInput	true	"Jeton de rafraîchissement"
//	@Success		204
//	@Failure		400	{object}	Problem
//	@Router			/api/v1/auth/logout [post]
func (h *Handler) logout(c *gin.Context) {
	var input domain.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	if err := h.service.Logout(c.Request.Context(), input); err != nil {
		handleServiceError(c, err, "Impossible de révoquer la session.")
		return
	}
	c.Status(http.StatusNoContent)
}

// getMe retourne le profil courant.
//
//	@Summary		Lire le profil courant
//	@Tags			Compte
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	UserResponse
//	@Failure		401	{object}	Problem
//	@Router			/api/v1/me [get]
func (h *Handler) getMe(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		return
	}
	user, err := h.service.GetMe(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err, "Impossible de charger le profil.")
		return
	}
	c.JSON(http.StatusOK, UserResponse{Data: user})
}

// patchMe modifie le profil courant.
//
//	@Summary		Modifier le profil courant
//	@Description	Les champs absents conservent leur valeur existante ; display_name: null efface le nom.
//	@Tags			Compte
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.UpdateProfileInput	true	"Champs à modifier"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Router			/api/v1/me [patch]
func (h *Handler) patchMe(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		return
	}
	var input domain.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	user, err := h.service.UpdateMe(c.Request.Context(), userID, input)
	if err != nil {
		handleServiceError(c, err, "Impossible de modifier le profil.")
		return
	}
	c.JSON(http.StatusOK, UserResponse{Data: user})
}

// deleteMe anonymise le compte courant :
//
//	@Summary		Supprimer le compte
//	@Description	Révoke les sessions et anonymise les données conservées à des fins statistiques.
//	@Tags			Compte
//	@Security		BearerAuth
//	@Success		204
//	@Failure		401	{object}	Problem
//	@Router			/api/v1/me [delete]
func (h *Handler) deleteMe(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		return
	}
	if err := h.service.DeleteMe(c.Request.Context(), userID); err != nil {
		handleServiceError(c, err, "Impossible de supprimer le compte.")
		return
	}
	c.Status(http.StatusNoContent)
}

// listMyContributions liste les observations soumises par l'utilisateur courant.
//
//	@Summary		Suivre ses contributions
//	@Description	Retourne les prix produits et carburants soumis, quel que soit leur statut.
//	@Tags			Compte
//	@Security		BearerAuth
//	@Produce		json
//	@Param			limit	query	integer	false	"Nombre de résultats" default(20) minimum(1) maximum(100)
//	@Param			offset	query	integer	false	"Décalage" default(0) minimum(0)
//	@Success		200		{object}	ContributionsResponse
//	@Failure		400		{object}	Problem
//	@Failure		401		{object}	Problem
//	@Router			/api/v1/me/contributions [get]
func (h *Handler) listMyContributions(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Un jeton d'accès est requis.")
		return
	}
	limit, offset, ok := pagination(c)
	if !ok {
		return
	}
	contributions, err := h.service.ListMyContributions(c.Request.Context(), userID, domain.Pagination{Limit: limit, Offset: offset})
	if err != nil {
		handleServiceError(c, err, "Impossible de charger les contributions.")
		return
	}
	c.JSON(http.StatusOK, ContributionsResponse{Data: contributions, Meta: PaginationMeta{Limit: limit, Offset: offset}})
}
