package httpapi

import (
	"errors"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
	"github.com/laurentpoirierfr/ojozone/internal/service"
)

const (
	productByIDRoute      = "/products/:id"
	productPriceByIDRoute = "/product-prices/:id"
	invalidJSONDetail     = "Le corps JSON est invalide."
)

type BuildInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

type Handler struct {
	service service.Service
	info    BuildInfo
}

type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type PaginationMeta struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ProductListResponse struct {
	Data []domain.Product `json:"data"`
	Meta PaginationMeta   `json:"meta"`
}

type ProductResponse struct {
	Data domain.Product `json:"data"`
}

type ProductPriceListResponse struct {
	Data []domain.ProductPrice `json:"data"`
	Meta PaginationMeta        `json:"meta"`
}

type ProductPriceResponse struct {
	Data domain.ProductPrice `json:"data"`
}

func NewRouter(appService service.Service, staticFS fs.FS, info BuildInfo) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	handler := &Handler{service: appService, info: info}

	ops := router.Group("/ops")
	ops.GET("/liveness", handler.liveness)
	ops.GET("/readiness", handler.readiness)
	ops.GET("/infos", handler.infos)

	api := router.Group("/api/v1")
	api.GET("/products", handler.listProducts)
	api.POST("/products", handler.upsertProduct)
	api.GET("/products/by-barcode/:barcode", handler.getProductByBarcode)
	api.GET(productByIDRoute, handler.getProduct)
	api.PUT(productByIDRoute, handler.replaceProduct)
	api.DELETE(productByIDRoute, handler.deleteProduct)
	api.GET("/products/:id/prices", handler.listProductPrices)
	api.GET("/product-prices", handler.listAllProductPrices)
	api.POST("/product-prices", handler.upsertProductPrice)
	api.GET(productPriceByIDRoute, handler.getProductPrice)
	api.PUT(productPriceByIDRoute, handler.replaceProductPrice)
	api.DELETE(productPriceByIDRoute, handler.deleteProductPrice)
	registerResourceRoutes(api, handler)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if staticFS != nil {
		fileSystem := http.FS(staticFS)
		index, err := fs.ReadFile(staticFS, "index.html")
		if err == nil {
			router.GET("/", func(c *gin.Context) { c.Data(http.StatusOK, "text/html; charset=utf-8", index) })
			router.NoRoute(staticFallback(fileSystem, index))
		}
	}
	return router
}

// liveness indique si le processus HTTP fonctionne.
//
//	@Summary		Vérifier la vitalité
//	@Description	Cette sonde ne dépend d'aucun service externe.
//	@Tags			Operations
//	@Produce		json
//	@Success		200	{object}	StatusResponse
//	@Router			/ops/liveness [get]
func (h *Handler) liveness(c *gin.Context) {
	c.JSON(http.StatusOK, StatusResponse{Status: "alive"})
}

// readiness indique si l'API peut traiter des requêtes.
//
//	@Summary		Vérifier la disponibilité
//	@Description	Vérifie notamment la connexion PostgreSQL.
//	@Tags			Operations
//	@Produce		json
//	@Success		200	{object}	StatusResponse
//	@Failure		503	{object}	Problem
//	@Router			/ops/readiness [get]
func (h *Handler) readiness(c *gin.Context) {
	if err := h.service.Readiness(c.Request.Context()); err != nil {
		writeProblem(c, http.StatusServiceUnavailable, "not_ready", "La base de données est indisponible.")
		return
	}
	c.JSON(http.StatusOK, StatusResponse{Status: "ready"})
}

// infos expose les métadonnées non sensibles du build.
//
//	@Summary	Obtenir les informations du service
//	@Tags		Operations
//	@Produce	json
//	@Success	200	{object}	BuildInfo
//	@Router		/ops/infos [get]
func (h *Handler) infos(c *gin.Context) {
	c.JSON(http.StatusOK, h.info)
}

// listProducts recherche les produits du catalogue.
//
//	@Summary	Lister les produits
//	@Tags		Produits
//	@Produce	json
//	@Param		q		query		string	false	"Nom, marque ou code-barres"
//	@Param		limit	query		integer	false	"Nombre de résultats" default(20) minimum(1) maximum(100)
//	@Param		offset	query		integer	false	"Décalage" default(0) minimum(0)
//	@Success	200		{object}	ProductListResponse
//	@Failure	400		{object}	Problem
//	@Failure	500		{object}	Problem
//	@Router		/api/v1/products [get]
func (h *Handler) listProducts(c *gin.Context) {
	filter, ok := productFilter(c)
	if !ok {
		return
	}
	products, err := h.service.ListProducts(c.Request.Context(), filter)
	if err != nil {
		handleServiceError(c, err, "Impossible de rechercher les produits.")
		return
	}
	c.JSON(http.StatusOK, ProductListResponse{Data: products, Meta: PaginationMeta{Limit: filter.Limit, Offset: filter.Offset}})
}

// upsertProduct crée ou remplace un produit selon sa clé métier.
//
//	@Summary	Créer ou mettre à jour un produit
//	@Description	Avec un code-barres, l'upsert utilise le code-barres. Sans code-barres, un UUID id est obligatoire.
//	@Tags		Produits
//	@Accept		json
//	@Produce	json
//	@Param		product	body		domain.ProductUpsert	true	"Produit à créer ou mettre à jour"
//	@Success	200		{object}	ProductResponse
//	@Failure	400		{object}	Problem
//	@Failure	409		{object}	Problem
//	@Failure	500		{object}	Problem
//	@Router		/api/v1/products [post]
func (h *Handler) upsertProduct(c *gin.Context) {
	var input domain.ProductUpsert
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	product, err := h.service.UpsertProduct(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, err, "Impossible d'enregistrer le produit.")
		return
	}
	c.JSON(http.StatusOK, ProductResponse{Data: product})
}

// getProduct retourne un produit par identifiant.
//
//	@Summary	Obtenir un produit
//	@Tags		Produits
//	@Produce	json
//	@Param		id	path		string	true	"UUID du produit" format(uuid)
//	@Success	200	{object}	ProductResponse
//	@Failure	400	{object}	Problem
//	@Failure	404	{object}	Problem
//	@Failure	500	{object}	Problem
//	@Router		/api/v1/products/{id} [get]
func (h *Handler) getProduct(c *gin.Context) {
	product, err := h.service.GetProduct(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "Impossible de charger le produit.")
		return
	}
	c.JSON(http.StatusOK, ProductResponse{Data: product})
}

// replaceProduct remplace un produit identifié par son UUID.
//
//	@Summary	Remplacer un produit
//	@Description	Crée le produit si l'UUID n'existe pas encore, sinon remplace ses champs modifiables.
//	@Tags		Produits
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string				true	"UUID du produit" format(uuid)
//	@Param		product	body		domain.ProductUpsert	true	"Nouvel état du produit"
//	@Success	200		{object}	ProductResponse
//	@Failure	400		{object}	Problem
//	@Failure	409		{object}	Problem
//	@Failure	500		{object}	Problem
//	@Router		/api/v1/products/{id} [put]
func (h *Handler) replaceProduct(c *gin.Context) {
	var input domain.ProductUpsert
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	product, err := h.service.ReplaceProduct(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		handleServiceError(c, err, "Impossible d'enregistrer le produit.")
		return
	}
	c.JSON(http.StatusOK, ProductResponse{Data: product})
}

// deleteProduct supprime un produit non référencé.
//
//	@Summary	Supprimer un produit
//	@Tags		Produits
//	@Param		id	path	string	true	"UUID du produit" format(uuid)
//	@Success	204
//	@Failure	400	{object}	Problem
//	@Failure	404	{object}	Problem
//	@Failure	409	{object}	Problem
//	@Failure	500	{object}	Problem
//	@Router		/api/v1/products/{id} [delete]
func (h *Handler) deleteProduct(c *gin.Context) {
	if err := h.service.DeleteProduct(c.Request.Context(), c.Param("id")); err != nil {
		handleServiceError(c, err, "Impossible de supprimer le produit.")
		return
	}
	c.Status(http.StatusNoContent)
}

// getProductByBarcode retourne un produit par code-barres.
//
//	@Summary	Obtenir un produit par code-barres
//	@Tags		Produits
//	@Produce	json
//	@Param		barcode	path		string	true	"Code-barres" maxlength(32)
//	@Success	200		{object}	ProductResponse
//	@Failure	400		{object}	Problem
//	@Failure	404		{object}	Problem
//	@Failure	500		{object}	Problem
//	@Router		/api/v1/products/by-barcode/{barcode} [get]
func (h *Handler) getProductByBarcode(c *gin.Context) {
	product, err := h.service.GetProductByBarcode(c.Request.Context(), c.Param("barcode"))
	if err != nil {
		handleServiceError(c, err, "Impossible de charger le produit.")
		return
	}
	c.JSON(http.StatusOK, ProductResponse{Data: product})
}

// listProductPrices retourne les observations de prix approuvées.
//
//	@Summary	Lister les prix d'un produit
//	@Tags		Prix
//	@Produce	json
//	@Param		id			path		string	true	"UUID du produit" format(uuid)
//	@Param		geo_area_id	query		string	false	"UUID de la zone géographique" format(uuid)
//	@Param		limit		query		integer	false	"Nombre de résultats" default(20) minimum(1) maximum(100)
//	@Param		offset		query		integer	false	"Décalage" default(0) minimum(0)
//	@Success	200			{object}	ProductPriceListResponse
//	@Failure	400			{object}	Problem
//	@Failure	500			{object}	Problem
//	@Router		/api/v1/products/{id}/prices [get]
func (h *Handler) listProductPrices(c *gin.Context) {
	filter, ok := priceFilter(c)
	if !ok {
		return
	}
	prices, err := h.service.ListProductPrices(c.Request.Context(), c.Param("id"), filter)
	if err != nil {
		handleServiceError(c, err, "Impossible de charger les prix.")
		return
	}
	c.JSON(http.StatusOK, ProductPriceListResponse{Data: prices, Meta: PaginationMeta{Limit: filter.Limit, Offset: filter.Offset}})
}

// listAllProductPrices retourne les observations de prix avec filtres optionnels.
//
//	@Summary	Lister toutes les observations de prix produit
//	@Tags		Prix
//	@Produce	json
//	@Param		product_id	query		string	false	"UUID du produit" format(uuid)
//	@Param		geo_area_id	query		string	false	"UUID de la zone géographique" format(uuid)
//	@Param		status		query		string	false	"Statut" Enums(pending,approved,rejected,flagged)
//	@Param		limit		query		integer	false	"Nombre de résultats" default(20) minimum(1) maximum(100)
//	@Param		offset		query		integer	false	"Décalage" default(0) minimum(0)
//	@Success	200			{object}	ProductPriceListResponse
//	@Failure	400			{object}	Problem
//	@Failure	500			{object}	Problem
//	@Router		/api/v1/product-prices [get]
func (h *Handler) listAllProductPrices(c *gin.Context) {
	limit, offset, ok := pagination(c)
	if !ok {
		return
	}
	filter := domain.PriceFilter{
		ProductID: strings.TrimSpace(c.Query("product_id")),
		GeoAreaID: strings.TrimSpace(c.Query("geo_area_id")),
		Status:    strings.TrimSpace(c.Query("status")),
		Limit:     limit,
		Offset:    offset,
	}
	prices, err := h.service.ListAllProductPrices(c.Request.Context(), filter)
	if err != nil {
		handleServiceError(c, err, "Impossible de charger les prix.")
		return
	}
	c.JSON(http.StatusOK, ProductPriceListResponse{Data: prices, Meta: PaginationMeta{Limit: limit, Offset: offset}})
}

// upsertProductPrice crée ou met à jour une observation de prix importée.
//
//	@Summary	Créer ou mettre à jour un prix produit
//	@Description	L'upsert utilise la clé composée source_id et source_record_id. source_record_id est donc obligatoire.
//	@Tags		Prix
//	@Accept		json
//	@Produce	json
//	@Param		price	body		domain.ProductPriceUpsert	true	"Observation de prix"
//	@Success	200		{object}	ProductPriceResponse
//	@Failure	400		{object}	Problem
//	@Failure	409		{object}	Problem
//	@Failure	500		{object}	Problem
//	@Router		/api/v1/product-prices [post]
func (h *Handler) upsertProductPrice(c *gin.Context) {
	var input domain.ProductPriceUpsert
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	price, err := h.service.UpsertProductPrice(c.Request.Context(), input)
	if err != nil {
		handleServiceError(c, err, "Impossible d'enregistrer le prix.")
		return
	}
	c.JSON(http.StatusOK, ProductPriceResponse{Data: price})
}

// getProductPrice retourne une observation de prix par UUID.
//
//	@Summary	Obtenir un prix produit
//	@Tags		Prix
//	@Produce	json
//	@Param		id	path		string	true	"UUID de l'observation" format(uuid)
//	@Success	200	{object}	ProductPriceResponse
//	@Failure	400	{object}	Problem
//	@Failure	404	{object}	Problem
//	@Failure	500	{object}	Problem
//	@Router		/api/v1/product-prices/{id} [get]
func (h *Handler) getProductPrice(c *gin.Context) {
	price, err := h.service.GetProductPrice(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleServiceError(c, err, "Impossible de charger le prix.")
		return
	}
	c.JSON(http.StatusOK, ProductPriceResponse{Data: price})
}

// replaceProductPrice remplace une observation de prix par UUID.
//
//	@Summary	Remplacer un prix produit
//	@Description	Crée l'observation si l'UUID n'existe pas, sinon remplace ses champs modifiables.
//	@Tags		Prix
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string					true	"UUID de l'observation" format(uuid)
//	@Param		price	body		domain.ProductPriceUpsert	true	"Nouvel état du prix"
//	@Success	200		{object}	ProductPriceResponse
//	@Failure	400		{object}	Problem
//	@Failure	409		{object}	Problem
//	@Failure	500		{object}	Problem
//	@Router		/api/v1/product-prices/{id} [put]
func (h *Handler) replaceProductPrice(c *gin.Context) {
	var input domain.ProductPriceUpsert
	if err := c.ShouldBindJSON(&input); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_json", invalidJSONDetail)
		return
	}
	price, err := h.service.ReplaceProductPrice(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		handleServiceError(c, err, "Impossible d'enregistrer le prix.")
		return
	}
	c.JSON(http.StatusOK, ProductPriceResponse{Data: price})
}

// deleteProductPrice supprime une observation de prix.
//
//	@Summary	Supprimer un prix produit
//	@Tags		Prix
//	@Param		id	path	string	true	"UUID de l'observation" format(uuid)
//	@Success	204
//	@Failure	400	{object}	Problem
//	@Failure	404	{object}	Problem
//	@Failure	500	{object}	Problem
//	@Router		/api/v1/product-prices/{id} [delete]
func (h *Handler) deleteProductPrice(c *gin.Context) {
	if err := h.service.DeleteProductPrice(c.Request.Context(), c.Param("id")); err != nil {
		handleServiceError(c, err, "Impossible de supprimer le prix.")
		return
	}
	c.Status(http.StatusNoContent)
}

func productFilter(c *gin.Context) (domain.ProductFilter, bool) {
	limit, offset, ok := pagination(c)
	return domain.ProductFilter{Search: c.Query("q"), Limit: limit, Offset: offset}, ok
}

func priceFilter(c *gin.Context) (domain.PriceFilter, bool) {
	limit, offset, ok := pagination(c)
	return domain.PriceFilter{GeoAreaID: strings.TrimSpace(c.Query("geo_area_id")), Limit: limit, Offset: offset}, ok
}

func pagination(c *gin.Context) (int32, int32, bool) {
	limit, err := strconv.ParseInt(c.DefaultQuery("limit", strconv.Itoa(service.DefaultPageSize)), 10, 32)
	if err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_pagination", "limit doit être un entier.")
		return 0, 0, false
	}
	offset, err := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)
	if err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_pagination", "offset doit être un entier.")
		return 0, 0, false
	}
	return int32(limit), int32(offset), true
}

func handleServiceError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeProblem(c, http.StatusNotFound, "not_found", "Ressource introuvable.")
	case errors.Is(err, domain.ErrConflict):
		writeProblem(c, http.StatusConflict, "conflict", "La ressource entre en conflit avec des données existantes.")
	case errors.Is(err, service.ErrInvalidID), errors.Is(err, service.ErrInvalidBarcode), errors.Is(err, service.ErrInvalidPagination), errors.Is(err, service.ErrInvalidProduct), errors.Is(err, service.ErrInvalidPrice), errors.Is(err, service.ErrInvalidResource):
		writeProblem(c, http.StatusBadRequest, "invalid_request", err.Error())
	default:
		writeProblem(c, http.StatusInternalServerError, "internal_error", fallback)
	}
}

func writeProblem(c *gin.Context, status int, problemType, detail string) {
	c.Header("Content-Type", "application/problem+json")
	c.JSON(status, Problem{Type: problemType, Title: http.StatusText(status), Status: status, Detail: detail})
}

func staticFallback(fileSystem http.FileSystem, index []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet || strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/ops/") {
			writeProblem(c, http.StatusNotFound, "not_found", "Ressource introuvable.")
			return
		}
		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if path != "" {
			if file, err := fileSystem.Open(path); err == nil {
				info, statErr := file.Stat()
				_ = file.Close()
				if statErr == nil && !info.IsDir() {
					c.FileFromFS(path, fileSystem)
					return
				}
			}
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	}
}
