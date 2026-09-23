package service

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/laurentpoirierfr/ojozone/internal/auth"
	"github.com/laurentpoirierfr/ojozone/internal/domain"
	"github.com/laurentpoirierfr/ojozone/internal/repository"
)

const (
	// DefaultPageSize est la taille de page appliquée en l'absence de paramètre.
	DefaultPageSize = 20
	// MaxPageSize limite le volume d'une réponse paginée.
	MaxPageSize = 100
)

var (
	// ErrInvalidID indique qu'un identifiant n'est pas un UUID valide.
	ErrInvalidID = errors.New("invalid identifier")
	// ErrInvalidBarcode indique qu'un code-barres est invalide.
	ErrInvalidBarcode = errors.New("invalid barcode")
	// ErrInvalidPagination indique des paramètres de pagination hors limites.
	ErrInvalidPagination = errors.New("invalid pagination")
	// ErrInvalidProduct indique un produit invalide.
	ErrInvalidProduct = errors.New("invalid product")
	// ErrInvalidPrice indique une observation de prix invalide.
	ErrInvalidPrice = errors.New("invalid product price")
	// ErrInvalidResource indique une charge utile invalide pour une ressource générique.
	ErrInvalidResource = errors.New("invalid resource")
	// ErrInvalidContribution indique une contribution invalide.
	ErrInvalidContribution = errors.New("invalid contribution")
	// ErrNoCommunitySource indique qu'aucune source citoyenne n'est configurée.
	ErrNoCommunitySource = errors.New("community source not configured")
	// ErrContributionNotModifiable indique qu'une contribution a déjà été traitée.
	ErrContributionNotModifiable = domain.ErrContributionNotModifiable
	// ErrInvalidModeration indique une décision de modération invalide.
	ErrInvalidModeration = errors.New("invalid moderation decision")
)

// Service expose les opérations métier consommées par les handlers HTTP.
type Service interface {
	Readiness(context.Context) error
	Register(context.Context, domain.RegisterInput) (domain.AuthResult, error)
	Login(context.Context, domain.LoginInput) (domain.AuthResult, error)
	Refresh(context.Context, domain.RefreshInput) (domain.AuthResult, error)
	Logout(context.Context, domain.RefreshInput) error
	GetMe(context.Context, string) (domain.User, error)
	UpdateMe(context.Context, string, domain.UpdateProfileInput) (domain.User, error)
	DeleteMe(context.Context, string) error
	ListMyContributions(context.Context, string, domain.Pagination) ([]domain.Contribution, error)
	SubmitProductContribution(context.Context, string, domain.ProductContributionInput) (domain.ContributionResult, error)
	SubmitFuelContribution(context.Context, string, domain.FuelContributionInput) (domain.ContributionResult, error)
	GetContribution(context.Context, string, string, string) (domain.ContributionDetail, error)
	UpdateContribution(context.Context, string, string, domain.ContributionPatch) (domain.ContributionDetail, error)
	DeleteContribution(context.Context, string, string) error
	ListModerationQueue(context.Context, domain.ModerationFilter) ([]domain.ModerationQueueItem, error)
	ReviewContribution(context.Context, string, string, domain.ContributionReviewInput) (domain.ContributionDetail, error)
	ListProducts(context.Context, domain.ProductFilter) ([]domain.Product, error)
	GetProduct(context.Context, string) (domain.Product, error)
	GetProductByBarcode(context.Context, string) (domain.Product, error)
	UpsertProduct(context.Context, domain.ProductUpsert) (domain.Product, error)
	ReplaceProduct(context.Context, string, domain.ProductUpsert) (domain.Product, error)
	DeleteProduct(context.Context, string) error
	ListProductPrices(context.Context, string, domain.PriceFilter) ([]domain.ProductPrice, error)
	ListAllProductPrices(context.Context, domain.PriceFilter) ([]domain.ProductPrice, error)
	GetProductPrice(context.Context, string) (domain.ProductPrice, error)
	UpsertProductPrice(context.Context, domain.ProductPriceUpsert) (domain.ProductPrice, error)
	ReplaceProductPrice(context.Context, string, domain.ProductPriceUpsert) (domain.ProductPrice, error)
	DeleteProductPrice(context.Context, string) error
	ListResource(context.Context, string, domain.Pagination) (any, error)
	GetResource(context.Context, string, domain.ResourceKey) (any, error)
	UpsertResource(context.Context, string, json.RawMessage) (any, error)
	ReplaceResource(context.Context, string, domain.ResourceKey, json.RawMessage) (any, error)
	DeleteResource(context.Context, string, domain.ResourceKey) error
}

// OjoZone implémente les règles métier et délègue la persistance au repository.
type OjoZone struct {
	repository repository.Repository
	auth       *auth.Manager
}

// New construit le service métier OjoZone.
func New(repository repository.Repository, manager *auth.Manager) *OjoZone {
	return &OjoZone{repository: repository, auth: manager}
}

// Readiness vérifie que les dépendances indispensables sont disponibles.
func (s *OjoZone) Readiness(ctx context.Context) error {
	checkContext, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return s.repository.Ping(checkContext)
}

// ListProducts recherche les produits avec pagination.
func (s *OjoZone) ListProducts(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	if err := validatePagination(filter.Limit, filter.Offset); err != nil {
		return nil, err
	}
	return s.repository.ListProducts(ctx, filter)
}

// GetProduct retourne un produit par UUID.
func (s *OjoZone) GetProduct(ctx context.Context, id string) (domain.Product, error) {
	if !validUUID(id) {
		return domain.Product{}, ErrInvalidID
	}
	return s.repository.GetProduct(ctx, id)
}

// GetProductByBarcode retourne un produit par code-barres.
func (s *OjoZone) GetProductByBarcode(ctx context.Context, barcode string) (domain.Product, error) {
	barcode = strings.TrimSpace(barcode)
	if barcode == "" || len(barcode) > 32 {
		return domain.Product{}, ErrInvalidBarcode
	}
	return s.repository.GetProductByBarcode(ctx, barcode)
}

// UpsertProduct crée ou met à jour un produit selon sa clé métier.
func (s *OjoZone) UpsertProduct(ctx context.Context, input domain.ProductUpsert) (domain.Product, error) {
	if err := normalizeAndValidateProduct(&input); err != nil {
		return domain.Product{}, err
	}
	if input.Barcode != nil {
		return s.repository.UpsertProductByBarcode(ctx, input)
	}
	if !validUUID(input.ID) {
		return domain.Product{}, ErrInvalidProduct
	}
	return s.repository.UpsertProductByID(ctx, input)
}

// ReplaceProduct crée ou remplace un produit par UUID.
func (s *OjoZone) ReplaceProduct(ctx context.Context, id string, input domain.ProductUpsert) (domain.Product, error) {
	if !validUUID(id) {
		return domain.Product{}, ErrInvalidID
	}
	input.ID = id
	if err := normalizeAndValidateProduct(&input); err != nil {
		return domain.Product{}, err
	}
	return s.repository.UpsertProductByID(ctx, input)
}

// DeleteProduct supprime un produit par UUID.
func (s *OjoZone) DeleteProduct(ctx context.Context, id string) error {
	if !validUUID(id) {
		return ErrInvalidID
	}
	return s.repository.DeleteProduct(ctx, id)
}

// ListProductPrices retourne les prix approuvés d'un produit.
func (s *OjoZone) ListProductPrices(ctx context.Context, productID string, filter domain.PriceFilter) ([]domain.ProductPrice, error) {
	if !validUUID(productID) || (filter.GeoAreaID != "" && !validUUID(filter.GeoAreaID)) {
		return nil, ErrInvalidID
	}
	if err := validatePagination(filter.Limit, filter.Offset); err != nil {
		return nil, err
	}
	return s.repository.ListProductPrices(ctx, productID, filter)
}

// ListAllProductPrices retourne les observations de prix avec filtres optionnels.
func (s *OjoZone) ListAllProductPrices(ctx context.Context, filter domain.PriceFilter) ([]domain.ProductPrice, error) {
	filter.ProductID = strings.TrimSpace(filter.ProductID)
	filter.GeoAreaID = strings.TrimSpace(filter.GeoAreaID)
	filter.Status = strings.TrimSpace(filter.Status)
	if filter.ProductID != "" && !validUUID(filter.ProductID) || filter.GeoAreaID != "" && !validUUID(filter.GeoAreaID) {
		return nil, ErrInvalidID
	}
	if filter.Status != "" && !validStatus(filter.Status) {
		return nil, ErrInvalidPrice
	}
	if err := validatePagination(filter.Limit, filter.Offset); err != nil {
		return nil, err
	}
	return s.repository.ListAllProductPrices(ctx, filter)
}

// GetProductPrice retourne une observation de prix par UUID.
func (s *OjoZone) GetProductPrice(ctx context.Context, id string) (domain.ProductPrice, error) {
	if !validUUID(id) {
		return domain.ProductPrice{}, ErrInvalidID
	}
	return s.repository.GetProductPrice(ctx, id)
}

// UpsertProductPrice crée ou met à jour un prix selon sa clé source.
func (s *OjoZone) UpsertProductPrice(ctx context.Context, input domain.ProductPriceUpsert) (domain.ProductPrice, error) {
	if err := normalizeAndValidateProductPrice(&input); err != nil || input.SourceRecordID == nil {
		return domain.ProductPrice{}, ErrInvalidPrice
	}
	return s.repository.UpsertProductPriceBySourceRecord(ctx, input)
}

// ReplaceProductPrice crée ou remplace un prix par UUID.
func (s *OjoZone) ReplaceProductPrice(ctx context.Context, id string, input domain.ProductPriceUpsert) (domain.ProductPrice, error) {
	if !validUUID(id) {
		return domain.ProductPrice{}, ErrInvalidID
	}
	input.ID = id
	if err := normalizeAndValidateProductPrice(&input); err != nil {
		return domain.ProductPrice{}, err
	}
	return s.repository.UpsertProductPriceByID(ctx, input)
}

// DeleteProductPrice supprime une observation de prix par UUID.
func (s *OjoZone) DeleteProductPrice(ctx context.Context, id string) error {
	if !validUUID(id) {
		return ErrInvalidID
	}
	return s.repository.DeleteProductPrice(ctx, id)
}

func validatePagination(limit, offset int32) error {
	if limit < 1 || limit > MaxPageSize || offset < 0 {
		return ErrInvalidPagination
	}
	return nil
}

func normalizeAndValidateProduct(input *domain.ProductUpsert) error {
	input.Name = strings.TrimSpace(input.Name)
	input.ReferenceUnit = strings.TrimSpace(input.ReferenceUnit)
	input.CategoryID = strings.TrimSpace(input.CategoryID)
	if input.Barcode != nil {
		barcode := strings.TrimSpace(*input.Barcode)
		input.Barcode = &barcode
	}
	if input.Name == "" || input.ReferenceUnit == "" || len(input.ReferenceUnit) > 16 || !validUUID(input.CategoryID) {
		return ErrInvalidProduct
	}
	if input.Barcode != nil && (*input.Barcode == "" || len(*input.Barcode) > 32) {
		return ErrInvalidProduct
	}
	if len(input.Attributes) == 0 {
		input.Attributes = json.RawMessage("{}")
	}
	if !json.Valid(input.Attributes) || input.Attributes[0] != '{' {
		return ErrInvalidProduct
	}
	return nil
}

func normalizeAndValidateProductPrice(input *domain.ProductPriceUpsert) error {
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.UnitCode = strings.TrimSpace(input.UnitCode)
	input.Status = strings.TrimSpace(input.Status)
	if input.SourceRecordID != nil {
		recordID := strings.TrimSpace(*input.SourceRecordID)
		input.SourceRecordID = &recordID
	}
	if !validUUID(input.ProductID) || !validUUID(input.LocationID) || !validUUID(input.SourceID) {
		return ErrInvalidPrice
	}
	if input.ContributorID != nil && !validUUID(*input.ContributorID) {
		return ErrInvalidPrice
	}
	if len(input.Currency) != 3 || input.UnitCode == "" || len(input.UnitCode) > 16 || input.ObservedAt.IsZero() {
		return ErrInvalidPrice
	}
	if !validNonNegativeDecimal(input.Amount) || !validPositiveDecimal(input.Quantity) || !validNonNegativeDecimal(input.NormalizedAmount) {
		return ErrInvalidPrice
	}
	if input.ConfidenceScore != nil && !validConfidence(*input.ConfidenceScore) {
		return ErrInvalidPrice
	}
	if input.SourceRecordID != nil && *input.SourceRecordID == "" {
		return ErrInvalidPrice
	}
	switch input.Status {
	case "pending", "approved", "rejected", "flagged":
	default:
		return ErrInvalidPrice
	}
	return nil
}

func validNonNegativeDecimal(value string) bool {
	number, ok := new(big.Rat).SetString(value)
	return ok && number.Sign() >= 0
}

func validPositiveDecimal(value string) bool {
	number, ok := new(big.Rat).SetString(value)
	return ok && number.Sign() > 0
}

func validConfidence(value string) bool {
	number, ok := new(big.Rat).SetString(value)
	return ok && number.Sign() >= 0 && number.Cmp(big.NewRat(1, 1)) <= 0
}

func validUUID(value string) bool {
	_, err := uuid.Parse(value)
	return err == nil
}
