package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/laurentpoirierfr/ojozone/internal/auth"
	"github.com/laurentpoirierfr/ojozone/internal/domain"
	"github.com/laurentpoirierfr/ojozone/internal/repository"
)

const (
	productID  = "ec2d9232-7ec5-44c4-85fc-1dfdd80b9d31"
	categoryID = "765c4c3e-9a2e-4a7f-a272-586a311cbb80"
	sessionID  = "9d2b4df0-3b6a-4a4f-bbfe-1c4e69f1b2a7"
	sourceID   = "0cbdc6bf-361b-4878-b407-e77f735098af"

	contributionID = "edfb822d-d78a-481e-9945-850e337a33fd"
)

func newTestService(repository repository.Repository) *OjoZone {
	return New(repository, auth.NewManager("test-secret-0123456789-abcdefghij", "ojozone-test", time.Minute))
}

type repositoryStub struct {
	upsertByIDCalls      int
	upsertByBarcodeCalls int
	lastInput            domain.ProductUpsert
	priceSourceCalls     int
	priceIDCalls         int
	lastPriceInput       domain.ProductPriceUpsert
	resourceCalls        int
	lastResource         string
	lastResourceInput    any
	lastResourceByID     bool
}

func (r *repositoryStub) Ping(context.Context) error { return nil }
func (r *repositoryStub) CreateUser(context.Context, domain.RegisterInput, string) (domain.User, error) {
	return domain.User{ID: productID, Email: "marie@example.com", Role: domain.RoleMember, Locale: "fr"}, nil
}
func (r *repositoryStub) GetUserByEmail(context.Context, string) (domain.UserWithPassword, error) {
	return domain.UserWithPassword{}, domain.ErrNotFound
}
func (r *repositoryStub) GetUserByID(context.Context, string) (domain.User, error) {
	return domain.User{ID: productID, Role: domain.RoleMember, Locale: "fr"}, nil
}
func (r *repositoryStub) UpdateUserProfile(context.Context, string, domain.UpdateProfileInput) (domain.User, error) {
	return domain.User{ID: productID, Role: domain.RoleMember, Locale: "fr"}, nil
}
func (r *repositoryStub) AnonymizeUser(context.Context, string) (domain.User, error) {
	return domain.User{ID: productID, Role: domain.RoleMember, Locale: "fr"}, nil
}
func (r *repositoryStub) CreateSession(context.Context, domain.Session) (domain.Session, error) {
	return domain.Session{ID: sessionID, UserID: productID}, nil
}
func (r *repositoryStub) GetSessionByRefreshHash(context.Context, string) (domain.Session, error) {
	return domain.Session{}, domain.ErrNotFound
}
func (r *repositoryStub) UpdateSessionRefresh(context.Context, string, string, time.Time) (domain.Session, error) {
	return domain.Session{ID: sessionID}, nil
}
func (r *repositoryStub) RevokeSession(context.Context, string) error { return nil }
func (r *repositoryStub) RevokeAllSessionsForUser(context.Context, string) error {
	return nil
}
func (r *repositoryStub) ListMyContributions(context.Context, string, domain.Pagination) ([]domain.Contribution, error) {
	return []domain.Contribution{}, nil
}
func (r *repositoryStub) FindCommunitySourceID(context.Context) (string, error) {
	return sourceID, nil
}
func (r *repositoryStub) CreateProductContribution(context.Context, domain.ProductContributionSubmit) (string, string, error) {
	return contributionID, domain.StatusPending, nil
}
func (r *repositoryStub) CreateFuelContribution(context.Context, domain.FuelContributionSubmit) (string, string, error) {
	return contributionID, domain.StatusPending, nil
}
func (r *repositoryStub) GetProductContribution(context.Context, string) (domain.ContributionDetail, error) {
	return domain.ContributionDetail{}, domain.ErrNotFound
}
func (r *repositoryStub) GetFuelContribution(context.Context, string) (domain.ContributionDetail, error) {
	return domain.ContributionDetail{}, domain.ErrNotFound
}
func (r *repositoryStub) UpdateProductContribution(context.Context, string, string, domain.ProductContributionSubmit) error {
	return nil
}
func (r *repositoryStub) UpdateFuelContribution(context.Context, string, string, domain.FuelContributionSubmit) error {
	return nil
}
func (r *repositoryStub) DeleteProductContribution(context.Context, string, string) error {
	return nil
}
func (r *repositoryStub) DeleteFuelContribution(context.Context, string, string) error {
	return nil
}
func (r *repositoryStub) ListProducts(context.Context, domain.ProductFilter) ([]domain.Product, error) {
	return nil, nil
}
func (r *repositoryStub) GetProduct(context.Context, string) (domain.Product, error) {
	return domain.Product{}, nil
}
func (r *repositoryStub) GetProductByBarcode(context.Context, string) (domain.Product, error) {
	return domain.Product{}, nil
}
func (r *repositoryStub) UpsertProductByID(_ context.Context, input domain.ProductUpsert) (domain.Product, error) {
	r.upsertByIDCalls++
	r.lastInput = input
	return domain.Product{ID: input.ID}, nil
}
func (r *repositoryStub) UpsertProductByBarcode(_ context.Context, input domain.ProductUpsert) (domain.Product, error) {
	r.upsertByBarcodeCalls++
	r.lastInput = input
	return domain.Product{}, nil
}
func (r *repositoryStub) DeleteProduct(context.Context, string) error { return nil }
func (r *repositoryStub) ListProductPrices(context.Context, string, domain.PriceFilter) ([]domain.ProductPrice, error) {
	return nil, nil
}
func (r *repositoryStub) ListAllProductPrices(context.Context, domain.PriceFilter) ([]domain.ProductPrice, error) {
	return nil, nil
}
func (r *repositoryStub) GetProductPrice(context.Context, string) (domain.ProductPrice, error) {
	return domain.ProductPrice{}, nil
}
func (r *repositoryStub) UpsertProductPriceBySourceRecord(context.Context, domain.ProductPriceUpsert) (domain.ProductPrice, error) {
	r.priceSourceCalls++
	return domain.ProductPrice{}, nil
}
func (r *repositoryStub) UpsertProductPriceByID(_ context.Context, input domain.ProductPriceUpsert) (domain.ProductPrice, error) {
	r.priceIDCalls++
	r.lastPriceInput = input
	return domain.ProductPrice{}, nil
}
func (r *repositoryStub) DeleteProductPrice(context.Context, string) error { return nil }
func (r *repositoryStub) ListModerationQueue(context.Context, *string, domain.Pagination) ([]domain.ModerationQueueItem, error) {
	return nil, nil
}
func (r *repositoryStub) ReviewContribution(context.Context, domain.ContributionReview) error {
	return nil
}
func (r *repositoryStub) ListResource(context.Context, string, domain.Pagination) (any, error) {
	return nil, nil
}
func (r *repositoryStub) GetResource(context.Context, string, domain.ResourceKey) (any, error) {
	return nil, nil
}
func (r *repositoryStub) UpsertResource(_ context.Context, resource string, input any, byID bool) (any, error) {
	r.resourceCalls++
	r.lastResource = resource
	r.lastResourceInput = input
	r.lastResourceByID = byID
	return input, nil
}
func (r *repositoryStub) DeleteResource(context.Context, string, domain.ResourceKey) error {
	return nil
}
func (r *repositoryStub) CreateImport(context.Context, domain.ImportCreate, string) (domain.Import, error) {
	return domain.Import{}, nil
}
func (r *repositoryStub) GetImport(context.Context, string) (domain.Import, error) {
	return domain.Import{}, domain.ErrNotFound
}
func (r *repositoryStub) ListImports(context.Context, domain.Pagination) ([]domain.Import, error) {
	return nil, nil
}
func (r *repositoryStub) ListImportRows(context.Context, string) ([]domain.ImportRow, error) {
	return nil, nil
}
func (r *repositoryStub) SetImportRowStatus(context.Context, string, int32, bool, *string) error {
	return nil
}
func (r *repositoryStub) MarkImportValidated(context.Context, string, int32, int32, []byte) error {
	return nil
}
func (r *repositoryStub) MarkImportPublished(context.Context, string, []byte) error {
	return nil
}

func TestUpsertProductUsesBarcodeWhenPresent(t *testing.T) {
	repository := &repositoryStub{}
	barcode := "3017620422003"
	_, err := newTestService(repository).UpsertProduct(context.Background(), validProductInput("", &barcode))
	if err != nil {
		t.Fatalf("upsert inattendu en erreur : %v", err)
	}
	if repository.upsertByBarcodeCalls != 1 || repository.upsertByIDCalls != 0 {
		t.Fatalf("appels inattendus : barcode=%d id=%d", repository.upsertByBarcodeCalls, repository.upsertByIDCalls)
	}
}

func TestUpsertProductWithoutBarcodeRequiresID(t *testing.T) {
	repository := &repositoryStub{}
	_, err := newTestService(repository).UpsertProduct(context.Background(), validProductInput("", nil))
	if err != ErrInvalidProduct {
		t.Fatalf("erreur obtenue %v, attendue %v", err, ErrInvalidProduct)
	}
	if repository.upsertByIDCalls != 0 {
		t.Fatal("le repository ne doit pas être appelé")
	}
}

func TestReplaceProductUsesPathID(t *testing.T) {
	repository := &repositoryStub{}
	input := validProductInput("b738785b-63a7-43b0-8133-a007ba314ed2", nil)
	_, err := newTestService(repository).ReplaceProduct(context.Background(), productID, input)
	if err != nil {
		t.Fatalf("remplacement inattendu en erreur : %v", err)
	}
	if repository.upsertByIDCalls != 1 || repository.lastInput.ID != productID {
		t.Fatalf("UUID transmis %q, attendu %q", repository.lastInput.ID, productID)
	}
}

func validProductInput(id string, barcode *string) domain.ProductUpsert {
	return domain.ProductUpsert{
		ID: id, CategoryID: categoryID, Name: "Lait", Barcode: barcode,
		ReferenceUnit: "L", Attributes: []byte(`{"fat_percent": 1.5}`),
	}
}

func TestUpsertProductPriceUsesSourceRecord(t *testing.T) {
	repository := &repositoryStub{}
	input := validPriceInput()
	_, err := newTestService(repository).UpsertProductPrice(context.Background(), input)
	if err != nil {
		t.Fatalf("upsert prix inattendu en erreur : %v", err)
	}
	if repository.priceSourceCalls != 1 || repository.priceIDCalls != 0 {
		t.Fatalf("appels inattendus : source=%d id=%d", repository.priceSourceCalls, repository.priceIDCalls)
	}
}

func TestUpsertProductPriceRequiresSourceRecord(t *testing.T) {
	repository := &repositoryStub{}
	input := validPriceInput()
	input.SourceRecordID = nil
	_, err := newTestService(repository).UpsertProductPrice(context.Background(), input)
	if err != ErrInvalidPrice {
		t.Fatalf("erreur obtenue %v, attendue %v", err, ErrInvalidPrice)
	}
}

func TestReplaceProductPriceUsesPathID(t *testing.T) {
	repository := &repositoryStub{}
	input := validPriceInput()
	_, err := newTestService(repository).ReplaceProductPrice(context.Background(), productID, input)
	if err != nil {
		t.Fatalf("remplacement prix inattendu en erreur : %v", err)
	}
	if repository.priceIDCalls != 1 || repository.lastPriceInput.ID != productID {
		t.Fatalf("UUID transmis %q, attendu %q", repository.lastPriceInput.ID, productID)
	}
}

func validPriceInput() domain.ProductPriceUpsert {
	recordID := "provider-row-1"
	return domain.ProductPriceUpsert{
		ProductID: productID, LocationID: "9a80c20f-9055-442d-97d1-43ee336230f0",
		SourceID: "0cbdc6bf-361b-4878-b407-e77f735098af", Amount: "2.35",
		Currency: "EUR", Quantity: "1.000", UnitCode: "L", NormalizedAmount: "2.3500",
		ObservedAt: time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC), Status: "approved",
		SourceRecordID: &recordID,
	}
}

func TestCategoryPostUsesSlugUpsert(t *testing.T) {
	repository := &repositoryStub{}
	_, err := newTestService(repository).UpsertResource(context.Background(), ResourceCategories, json.RawMessage(`{"slug":"food","name_i18n":{"fr":"Alimentation"}}`))
	if err != nil {
		t.Fatalf("upsert catégorie inattendu en erreur : %v", err)
	}
	input, ok := repository.lastResourceInput.(domain.CategoryUpsert)
	if repository.lastResource != ResourceCategories || repository.lastResourceByID || !ok || input.Slug != "food" {
		t.Fatalf("mauvaise clé d'upsert : %#v", repository)
	}
}

func TestFuelPricePostRequiresAndUsesSourceRecord(t *testing.T) {
	repository := &repositoryStub{}
	raw := json.RawMessage(`{"fuel_type_id":"ec2d9232-7ec5-44c4-85fc-1dfdd80b9d31","location_id":"9a80c20f-9055-442d-97d1-43ee336230f0","source_id":"0cbdc6bf-361b-4878-b407-e77f735098af","amount_per_litre":"1.8990","currency":"EUR","observed_at":"2026-09-22T12:00:00Z","status":"approved","source_record_id":"fuel-42"}`)
	_, err := newTestService(repository).UpsertResource(context.Background(), ResourceFuelPrices, raw)
	if err != nil {
		t.Fatalf("upsert carburant inattendu en erreur : %v", err)
	}
	input, ok := repository.lastResourceInput.(domain.FuelPriceUpsert)
	if repository.lastResourceByID || !ok || input.SourceRecordID == nil || *input.SourceRecordID != "fuel-42" {
		t.Fatalf("source_record non transmis : %#v", repository.lastResourceInput)
	}
}
