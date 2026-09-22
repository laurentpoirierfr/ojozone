package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/laurentpoirierfr/ojozone/internal/auth"
	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

const (
	authUserID    = "9b1c3e2f-2a6c-4f5a-8e9d-1b2c3d4e5f6a"
	authSessionID = "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
)

type authRepositoryStub struct {
	userWithPassword domain.UserWithPassword
	user             domain.User
	getByEmailError  error
	createUserError  error
	session          domain.Session
	anonError        error
	updateHash       string
	updateCalled     bool
}

func (r *authRepositoryStub) Ping(context.Context) error { return nil }
func (r *authRepositoryStub) CreateUser(context.Context, domain.RegisterInput, string) (domain.User, error) {
	if r.createUserError != nil {
		return domain.User{}, r.createUserError
	}
	return r.user, nil
}
func (r *authRepositoryStub) GetUserByEmail(context.Context, string) (domain.UserWithPassword, error) {
	return r.userWithPassword, r.getByEmailError
}
func (r *authRepositoryStub) GetUserByID(context.Context, string) (domain.User, error) {
	return r.user, nil
}
func (r *authRepositoryStub) UpdateUserProfile(context.Context, string, domain.UpdateProfileInput) (domain.User, error) {
	return r.user, nil
}
func (r *authRepositoryStub) AnonymizeUser(context.Context, string) (domain.User, error) {
	if r.anonError != nil {
		return domain.User{}, r.anonError
	}
	return r.user, nil
}
func (r *authRepositoryStub) CreateSession(context.Context, domain.Session) (domain.Session, error) {
	return r.session, nil
}
func (r *authRepositoryStub) GetSessionByRefreshHash(context.Context, string) (domain.Session, error) {
	return r.session, nil
}
func (r *authRepositoryStub) UpdateSessionRefresh(_ context.Context, _ string, hash string, _ time.Time) (domain.Session, error) {
	r.updateCalled = true
	r.updateHash = hash
	return r.session, nil
}
func (r *authRepositoryStub) RevokeSession(context.Context, string) error { return nil }
func (r *authRepositoryStub) RevokeAllSessionsForUser(context.Context, string) error {
	return nil
}
func (r *authRepositoryStub) ListMyContributions(context.Context, string, domain.Pagination) ([]domain.Contribution, error) {
	return []domain.Contribution{}, nil
}
func (r *authRepositoryStub) FindCommunitySourceID(context.Context) (string, error) {
	return sourceID, nil
}
func (r *authRepositoryStub) CreateProductContribution(context.Context, domain.ProductContributionSubmit) (string, string, error) {
	return contributionID, domain.StatusPending, nil
}
func (r *authRepositoryStub) CreateFuelContribution(context.Context, domain.FuelContributionSubmit) (string, string, error) {
	return contributionID, domain.StatusPending, nil
}
func (r *authRepositoryStub) GetProductContribution(context.Context, string) (domain.ContributionDetail, error) {
	return domain.ContributionDetail{}, domain.ErrNotFound
}
func (r *authRepositoryStub) GetFuelContribution(context.Context, string) (domain.ContributionDetail, error) {
	return domain.ContributionDetail{}, domain.ErrNotFound
}
func (r *authRepositoryStub) UpdateProductContribution(context.Context, string, string, domain.ProductContributionSubmit) error {
	return nil
}
func (r *authRepositoryStub) UpdateFuelContribution(context.Context, string, string, domain.FuelContributionSubmit) error {
	return nil
}
func (r *authRepositoryStub) DeleteProductContribution(context.Context, string, string) error {
	return nil
}
func (r *authRepositoryStub) DeleteFuelContribution(context.Context, string, string) error {
	return nil
}
func (r *authRepositoryStub) ListProducts(context.Context, domain.ProductFilter) ([]domain.Product, error) {
	return nil, nil
}
func (r *authRepositoryStub) GetProduct(context.Context, string) (domain.Product, error) {
	return domain.Product{}, nil
}
func (r *authRepositoryStub) GetProductByBarcode(context.Context, string) (domain.Product, error) {
	return domain.Product{}, nil
}
func (r *authRepositoryStub) UpsertProductByID(context.Context, domain.ProductUpsert) (domain.Product, error) {
	return domain.Product{}, nil
}
func (r *authRepositoryStub) UpsertProductByBarcode(context.Context, domain.ProductUpsert) (domain.Product, error) {
	return domain.Product{}, nil
}
func (r *authRepositoryStub) DeleteProduct(context.Context, string) error { return nil }
func (r *authRepositoryStub) ListProductPrices(context.Context, string, domain.PriceFilter) ([]domain.ProductPrice, error) {
	return nil, nil
}
func (r *authRepositoryStub) ListAllProductPrices(context.Context, domain.PriceFilter) ([]domain.ProductPrice, error) {
	return nil, nil
}
func (r *authRepositoryStub) GetProductPrice(context.Context, string) (domain.ProductPrice, error) {
	return domain.ProductPrice{}, nil
}
func (r *authRepositoryStub) UpsertProductPriceBySourceRecord(context.Context, domain.ProductPriceUpsert) (domain.ProductPrice, error) {
	return domain.ProductPrice{}, nil
}
func (r *authRepositoryStub) UpsertProductPriceByID(context.Context, domain.ProductPriceUpsert) (domain.ProductPrice, error) {
	return domain.ProductPrice{}, nil
}
func (r *authRepositoryStub) DeleteProductPrice(context.Context, string) error { return nil }
func (r *authRepositoryStub) ListResource(context.Context, string, domain.Pagination) (any, error) {
	return nil, nil
}
func (r *authRepositoryStub) GetResource(context.Context, string, domain.ResourceKey) (any, error) {
	return nil, nil
}
func (r *authRepositoryStub) UpsertResource(context.Context, string, any, bool) (any, error) {
	return nil, nil
}
func (r *authRepositoryStub) DeleteResource(context.Context, string, domain.ResourceKey) error {
	return nil
}

func authTestRepository() *authRepositoryStub {
	hash, _ := auth.HashPassword("correct horse battery")
	return &authRepositoryStub{
		user: domain.User{ID: authUserID, Email: "marie@example.com", Role: domain.RoleMember, Locale: "fr"},
		userWithPassword: domain.UserWithPassword{
			User:         domain.User{ID: authUserID, Email: "marie@example.com", Role: domain.RoleMember, Locale: "fr"},
			PasswordHash: hash,
		},
		session: domain.Session{ID: authSessionID, UserID: authUserID, ExpiresAt: time.Now().Add(RefreshTokenTTL)},
	}
}

func TestRegisterValidatesFieldsBeforeRepository(t *testing.T) {
	repository := authTestRepository()
	service := newTestService(repository)

	if _, err := service.Register(context.Background(), domain.RegisterInput{Email: "pas-un-email", Password: "short", Locale: "fr"}); !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("erreur obtenue %v, attendue %v", err, ErrInvalidProfile)
	}
	if _, err := service.Register(context.Background(), domain.RegisterInput{Email: "marie@example.com", Password: "valide-123", Locale: "de"}); !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("locale : erreur obtenue %v, attendue %v", err, ErrInvalidProfile)
	}
}

func TestRegisterIssuesTokenPair(t *testing.T) {
	service := newTestService(authTestRepository())
	result, err := service.Register(context.Background(), domain.RegisterInput{Email: "Marie@Example.com", Password: "valide-123", Locale: "fr"})
	if err != nil {
		t.Fatalf("inscription inattendue en erreur : %v", err)
	}
	if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" || result.Tokens.TokenType != "Bearer" {
		t.Fatal("paire de jetons incomplète")
	}
	if result.User.Email != "marie@example.com" {
		t.Fatalf("email non normalisé : %q", result.User.Email)
	}
	parsed, err := service.auth.ParseAccessToken(result.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("jeton d'accès non lisible : %v", err)
	}
	if parsed.Subject != authUserID || parsed.Role != domain.RoleMember {
		t.Fatalf("revendications inattendues : subject=%s role=%s", parsed.Subject, parsed.Role)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	repository := authTestRepository()
	repository.createUserError = domain.ErrConflict
	service := newTestService(repository)
	_, err := service.Register(context.Background(), domain.RegisterInput{Email: "marie@example.com", Password: "valide-123", Locale: "fr"})
	if !errors.Is(err, ErrEmailAlreadyUsed) {
		t.Fatalf("erreur obtenue %v, attendue %v", err, ErrEmailAlreadyUsed)
	}
}

func TestLoginRejectsUnknownUser(t *testing.T) {
	repository := authTestRepository()
	repository.getByEmailError = domain.ErrNotFound
	service := newTestService(repository)
	_, err := service.Login(context.Background(), domain.LoginInput{Email: "inconnu@example.com", Password: "whatever-123"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("erreur obtenue %v, attendue %v", err, ErrInvalidCredentials)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	service := newTestService(authTestRepository())
	_, err := service.Login(context.Background(), domain.LoginInput{Email: "marie@example.com", Password: "mauvais-mot-de-passe"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("erreur obtenue %v, attendue %v", err, ErrInvalidCredentials)
	}
}

func TestRefreshRotatesRefreshToken(t *testing.T) {
	repository := authTestRepository()
	service := newTestService(repository)
	previous, _, _ := service.auth.NewRefreshToken()
	_, loginHash, _ := service.auth.NewRefreshToken()
	repository.session.RefreshTokenHash = loginHash

	result, err := service.Refresh(context.Background(), domain.RefreshInput{RefreshToken: previous})
	if err != nil {
		t.Fatalf("renouvellement inattendu en erreur : %v", err)
	}
	if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" {
		t.Fatal("paire de jetons incomplète")
	}
	if result.Tokens.RefreshToken == previous {
		t.Fatal("le jeton de rafraîchissement doit être remplacé")
	}
	if !repository.updateCalled {
		t.Fatal("la rotation en base doit être demandée")
	}
	if repository.updateHash == "" {
		t.Fatal("le nouvel hash doit être persisté")
	}
}

func TestRefreshRejectsRevokedSession(t *testing.T) {
	repository := authTestRepository()
	revokedAt := time.Now()
	repository.session.RevokedAt = &revokedAt
	service := newTestService(repository)
	_, err := service.Refresh(context.Background(), domain.RefreshInput{RefreshToken: "some-raw-token"})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("erreur obtenue %v, attendue %v", err, domain.ErrUnauthorized)
	}
}

func TestDeleteMeAnonymizesAndRevokesSessions(t *testing.T) {
	repository := authTestRepository()
	service := newTestService(repository)
	if err := service.DeleteMe(context.Background(), authUserID); err != nil {
		t.Fatalf("suppression inattendue en erreur : %v", err)
	}
}

func TestUpdateMeValidatesProfile(t *testing.T) {
	service := newTestService(authTestRepository())
	if _, err := service.UpdateMe(context.Background(), authUserID, domain.UpdateProfileInput{Locale: stringPtr("de")}); !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("erreur obtenue %v, attendue %v", err, ErrInvalidProfile)
	}
}

func stringPtr(value string) *string { return &value }
