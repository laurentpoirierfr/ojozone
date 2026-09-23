package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/laurentpoirierfr/ojozone/internal/auth"
	"github.com/laurentpoirierfr/ojozone/internal/domain"
	"github.com/laurentpoirierfr/ojozone/internal/repository"
)

const (
	contributorUserID = "9b1c3e2f-2a6c-4f5a-8e9d-1b2c3d4e5f6a"
	otherUserID       = "7c4e8a2b-6d1f-4a9b-8e5c-3f2a1b0c9d8e"
)

type contributionsRepositoryStub struct {
	repository.Repository

	communitySourceID string
	communityError    error
	communityNotFound bool

	productDetail domain.ContributionDetail
	fuelDetail    domain.ContributionDetail

	createdID     string
	createdStatus string
	createError   error

	productSubmit domain.ProductContributionSubmit
	fuelSubmit    domain.FuelContributionSubmit
	productUpdate domain.ProductContributionSubmit
	fuelUpdate    domain.FuelContributionSubmit
	updateError   error

	deleteError error

	review      domain.ContributionReview
	reviewError error
	queueItems  []domain.ModerationQueueItem
	queueStatus *string
}

func (s *contributionsRepositoryStub) FindCommunitySourceID(context.Context) (string, error) {
	if s.communityNotFound {
		return "", domain.ErrNotFound
	}
	return s.communitySourceID, s.communityError
}

func (s *contributionsRepositoryStub) CreateProductContribution(_ context.Context, submit domain.ProductContributionSubmit) (string, string, error) {
	s.productSubmit = submit
	if s.createError != nil {
		return "", "", s.createError
	}
	return s.createdID, s.createdStatus, nil
}

func (s *contributionsRepositoryStub) CreateFuelContribution(_ context.Context, submit domain.FuelContributionSubmit) (string, string, error) {
	s.fuelSubmit = submit
	if s.createError != nil {
		return "", "", s.createError
	}
	return s.createdID, s.createdStatus, nil
}

func (s *contributionsRepositoryStub) GetProductContribution(context.Context, string) (domain.ContributionDetail, error) {
	if s.productDetail.ID == "" {
		return domain.ContributionDetail{}, domain.ErrNotFound
	}
	return s.productDetail, nil
}

func (s *contributionsRepositoryStub) GetFuelContribution(context.Context, string) (domain.ContributionDetail, error) {
	if s.fuelDetail.ID == "" {
		return domain.ContributionDetail{}, domain.ErrNotFound
	}
	return s.fuelDetail, nil
}

func (s *contributionsRepositoryStub) UpdateProductContribution(_ context.Context, _, _ string, submit domain.ProductContributionSubmit) error {
	s.productUpdate = submit
	if s.updateError != nil {
		return s.updateError
	}
	s.productDetail.Contribution.Amount = submit.Amount
	if submit.Quantity != "" {
		quantity := submit.Quantity
		s.productDetail.Quantity = &quantity
	}
	if submit.UnitCode != "" {
		unit := submit.UnitCode
		s.productDetail.UnitCode = &unit
	}
	return nil
}

func (s *contributionsRepositoryStub) UpdateFuelContribution(_ context.Context, _, _ string, submit domain.FuelContributionSubmit) error {
	s.fuelUpdate = submit
	if s.updateError != nil {
		return s.updateError
	}
	s.fuelDetail.Contribution.Amount = submit.AmountPerLitre
	return nil
}

func (s *contributionsRepositoryStub) DeleteProductContribution(context.Context, string, string) error {
	return s.deleteError
}

func (s *contributionsRepositoryStub) DeleteFuelContribution(context.Context, string, string) error {
	return s.deleteError
}

func (s *contributionsRepositoryStub) ListModerationQueue(_ context.Context, status *string, _ domain.Pagination) ([]domain.ModerationQueueItem, error) {
	s.queueStatus = status
	return s.queueItems, nil
}

func (s *contributionsRepositoryStub) ReviewContribution(_ context.Context, review domain.ContributionReview) error {
	s.review = review
	if s.reviewError != nil {
		return s.reviewError
	}
	switch review.Type {
	case "product_price":
		s.productDetail.Contribution.Status = review.NewStatus
	case "fuel_price":
		s.fuelDetail.Contribution.Status = review.NewStatus
	}
	return nil
}

func newContribTestService(stub *contributionsRepositoryStub) *OjoZone {
	if stub == nil {
		stub = &contributionsRepositoryStub{}
	}
	return New(stub, auth.NewManager("contrib-test-secret-0123456789-abcdefghij", "ojozone-test", time.Minute))
}

func testObservedAt() time.Time {
	return time.Date(2026, 9, 22, 17, 30, 0, 0, time.UTC)
}

func productContributionInput() domain.ProductContributionInput {
	return domain.ProductContributionInput{
		ProductID:  "ec2d9232-7ec5-44c4-85fc-1dfdd80b9d31",
		LocationID: "9a80c20f-9055-442d-97d1-43ee336230f0",
		Amount:     "2.35",
		Currency:   "EUR",
		Quantity:   "1.000",
		UnitCode:   "L",
		ObservedAt: testObservedAt(),
	}
}

func fuelContributionInput() domain.FuelContributionInput {
	return domain.FuelContributionInput{
		FuelTypeID:     "b25fbc0e-9d7a-4d2e-bf5f-7d2a1f0a3c92",
		LocationID:     "9a80c20f-9055-442d-97d1-43ee336230f0",
		AmountPerLitre: "1.849",
		Currency:       "EUR",
		ObservedAt:     testObservedAt(),
	}
}

func productContributionDetail() domain.ContributionDetail {
	return domain.ContributionDetail{
		Contribution: domain.Contribution{
			ID: contributionID, Type: "product_price", Status: domain.StatusPending,
			ObservedAt: testObservedAt(), CreatedAt: testObservedAt(),
			Currency: "EUR", Amount: "2.35",
		},
		ContributorID: contributorUserID,
	}
}

func fuelContributionDetail() domain.ContributionDetail {
	return domain.ContributionDetail{
		Contribution: domain.Contribution{
			ID: contributionID, Type: "fuel_price", Status: domain.StatusPending,
			ObservedAt: testObservedAt(), CreatedAt: testObservedAt(),
			Currency: "EUR", Amount: "1.849",
		},
		ContributorID: contributorUserID,
	}
}

func TestSubmitProductContribution(t *testing.T) {
	stub := &contributionsRepositoryStub{communitySourceID: sourceID, createdID: contributionID, createdStatus: domain.StatusPending}
	assertEmpty, err := newContribTestService(stub).SubmitProductContribution(context.Background(), contributorUserID, productContributionInput())
	if err != nil {
		t.Fatalf("soumission : %v", err)
	}
	if assertEmpty.ID != contributionID || assertEmpty.Status != domain.StatusPending {
		t.Fatalf("resultat inattendu : %+v", assertEmpty)
	}
	if len(assertEmpty.Checks) != 1 || assertEmpty.Checks[0].Code != "format_valid" {
		t.Fatalf("contrôles inattendus : %+v", assertEmpty.Checks)
	}
}

func TestSubmitProductContributionNormalizesAmount(t *testing.T) {
	stub := &contributionsRepositoryStub{communitySourceID: sourceID, createdID: contributionID, createdStatus: domain.StatusPending}
	service := newContribTestService(stub)
	input := productContributionInput()
	input.Amount, input.Quantity = "2.35", "2.000"
	if _, err := service.SubmitProductContribution(context.Background(), contributorUserID, input); err != nil {
		t.Fatalf("soumission : %v", err)
	}
	if stub.productSubmit.NormalizedAmount != "1.1750" {
		t.Fatalf("montant normalisé attendu 1.1750, obtenu : %q", stub.productSubmit.NormalizedAmount)
	}
	if stub.productSubmit.ContributorID != contributorUserID || stub.productSubmit.SourceID != sourceID {
		t.Fatalf("auteur ou source inattendus : %+v", stub.productSubmit)
	}
}

func TestSubmitProductContributionRejectsInvalid(t *testing.T) {
	service := newContribTestService(&contributionsRepositoryStub{communitySourceID: sourceID})
	cases := []struct {
		name   string
		mutate func(*domain.ProductContributionInput)
	}{
		{"produit inconnu", func(i *domain.ProductContributionInput) { i.ProductID = "pas-un-uuid" }},
		{"lieu inconnu", func(i *domain.ProductContributionInput) { i.LocationID = "pas-un-uuid" }},
		{"montant negatif", func(i *domain.ProductContributionInput) { i.Amount = "-1.00" }},
		{"quantite nulle", func(i *domain.ProductContributionInput) { i.Quantity = "0" }},
		{"unite vide", func(i *domain.ProductContributionInput) { i.UnitCode = "" }},
		{"devise courte", func(i *domain.ProductContributionInput) { i.Currency = "EU" }},
		{"date absente", func(i *domain.ProductContributionInput) { i.ObservedAt = time.Time{} }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := productContributionInput()
			test.mutate(&input)
			if _, err := service.SubmitProductContribution(context.Background(), contributorUserID, input); !errors.Is(err, ErrInvalidContribution) {
				t.Fatalf("erreur attendue %v, obtenu : %v", ErrInvalidContribution, err)
			}
		})
	}
}

func TestSubmitFuelContribution(t *testing.T) {
	stub := &contributionsRepositoryStub{communitySourceID: sourceID, createdID: contributionID, createdStatus: domain.StatusPending}
	result, err := newContribTestService(stub).SubmitFuelContribution(context.Background(), contributorUserID, fuelContributionInput())
	if err != nil {
		t.Fatalf("soumission : %v", err)
	}
	if result.ID != contributionID || result.Status != domain.StatusPending {
		t.Fatalf("resultat inattendu : %+v", result)
	}
}

func TestSubmitFuelContributionRejectsInvalid(t *testing.T) {
	service := newContribTestService(&contributionsRepositoryStub{communitySourceID: sourceID})
	input := fuelContributionInput()
	input.AmountPerLitre = "0"
	if _, err := service.SubmitFuelContribution(context.Background(), contributorUserID, input); !errors.Is(err, ErrInvalidContribution) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrInvalidContribution, err)
	}
}

func TestSubmitContributionWithoutCommunitySource(t *testing.T) {
	stub := &contributionsRepositoryStub{communityNotFound: true}
	service := newContribTestService(stub)
	if _, err := service.SubmitProductContribution(context.Background(), contributorUserID, productContributionInput()); !errors.Is(err, ErrNoCommunitySource) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrNoCommunitySource, err)
	}
	if _, err := service.SubmitFuelContribution(context.Background(), contributorUserID, fuelContributionInput()); !errors.Is(err, ErrNoCommunitySource) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrNoCommunitySource, err)
	}
}

func TestGetContributionByOwner(t *testing.T) {
	stub := &contributionsRepositoryStub{fuelDetail: fuelContributionDetail()}
	service := newContribTestService(stub)
	detail, err := service.GetContribution(context.Background(), contributorUserID, domain.RoleMember, contributionID)
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if detail.ID != contributionID || detail.Type != "fuel_price" {
		t.Fatalf("detail inattendu : %+v", detail)
	}
}

func TestGetContributionByModerator(t *testing.T) {
	stub := &contributionsRepositoryStub{productDetail: productContributionDetail()}
	service := newContribTestService(stub)
	detail, err := service.GetContribution(context.Background(), otherUserID, domain.RoleModerator, contributionID)
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if detail.Type != "product_price" {
		t.Fatalf("type inattendu : %+v", detail)
	}
}

func TestGetContributionForbiddenForOthers(t *testing.T) {
	stub := &contributionsRepositoryStub{productDetail: productContributionDetail()}
	service := newContribTestService(stub)
	if _, err := service.GetContribution(context.Background(), otherUserID, domain.RoleMember, contributionID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("erreur attendue %v, obtenu : %v", domain.ErrForbidden, err)
	}
}

func TestGetContributionNotFound(t *testing.T) {
	service := newContribTestService(nil)
	if _, err := service.GetContribution(context.Background(), contributorUserID, domain.RoleMember, contributionID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("erreur attendue %v, obtenu : %v", domain.ErrNotFound, err)
	}
}

func TestUpdateContributionProduct(t *testing.T) {
	detail := productContributionDetail()
	quantity := "1.000"
	unit := "L"
	detail.Quantity = &quantity
	detail.UnitCode = &unit
	stub := &contributionsRepositoryStub{productDetail: detail}
	service := newContribTestService(stub)
	amount := "2.75"
	updated, err := service.UpdateContribution(context.Background(), contributorUserID, contributionID, domain.ContributionPatch{Amount: &amount})
	if err != nil {
		t.Fatalf("correction : %v", err)
	}
	if updated.Amount != "2.75" {
		t.Fatalf("montant non corrigé : %+v", updated)
	}
	if stub.productUpdate.Amount != "2.75" || stub.productUpdate.NormalizedAmount != "2.7500" {
		t.Fatalf("champs transmis inattendus : %+v", stub.productUpdate)
	}
}

func TestUpdateContributionForbidden(t *testing.T) {
	stub := &contributionsRepositoryStub{productDetail: productContributionDetail()}
	service := newContribTestService(stub)
	if _, err := service.UpdateContribution(context.Background(), otherUserID, contributionID, domain.ContributionPatch{}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("erreur attendue %v, obtenu : %v", domain.ErrForbidden, err)
	}
}

func TestUpdateContributionNotModifiable(t *testing.T) {
	detail := productContributionDetail()
	detail.Status = domain.StatusApproved
	stub := &contributionsRepositoryStub{productDetail: detail}
	service := newContribTestService(stub)
	if _, err := service.UpdateContribution(context.Background(), contributorUserID, contributionID, domain.ContributionPatch{}); !errors.Is(err, ErrContributionNotModifiable) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrContributionNotModifiable, err)
	}
}

func TestDeleteContribution(t *testing.T) {
	stub := &contributionsRepositoryStub{fuelDetail: fuelContributionDetail()}
	service := newContribTestService(stub)
	if err := service.DeleteContribution(context.Background(), contributorUserID, contributionID); err != nil {
		t.Fatalf("retrait : %v", err)
	}
}

func TestDeleteContributionForbidden(t *testing.T) {
	stub := &contributionsRepositoryStub{fuelDetail: fuelContributionDetail()}
	service := newContribTestService(stub)
	if err := service.DeleteContribution(context.Background(), otherUserID, contributionID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("erreur attendue %v, obtenu : %v", domain.ErrForbidden, err)
	}
}

func TestDeleteContributionNotModifiable(t *testing.T) {
	detail := fuelContributionDetail()
	detail.Status = domain.StatusRejected
	stub := &contributionsRepositoryStub{fuelDetail: detail}
	service := newContribTestService(stub)
	if err := service.DeleteContribution(context.Background(), contributorUserID, contributionID); !errors.Is(err, ErrContributionNotModifiable) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrContributionNotModifiable, err)
	}
}
