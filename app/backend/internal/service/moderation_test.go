package service

import (
	"context"
	"errors"
	"testing"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

func approveInput(note string) domain.ContributionReviewInput {
	return domain.ContributionReviewInput{Decision: domain.StatusApproved, Note: strptr(note)}
}

func rejectInput(note string) domain.ContributionReviewInput {
	return domain.ContributionReviewInput{Decision: domain.StatusRejected, Note: strptr(note)}
}

func strptr(value string) *string {
	text := value
	return &text
}

func TestListModerationQueueDefaultsToAllStatuses(t *testing.T) {
	stub := &contributionsRepositoryStub{queueItems: []domain.ModerationQueueItem{{ID: contributionID}}}
	items, err := newContribTestService(stub).ListModerationQueue(context.Background(), domain.ModerationFilter{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("file : %v", err)
	}
	if len(items) != 1 || stub.queueStatus != nil {
		t.Fatalf("file inattendue : %+v, statut %v", items, stub.queueStatus)
	}
}

func TestListModerationQueueFiltersByStatus(t *testing.T) {
	stub := &contributionsRepositoryStub{}
	status := "pending"
	if _, err := newContribTestService(stub).ListModerationQueue(context.Background(), domain.ModerationFilter{Status: status, Limit: 20, Offset: 0}); err != nil {
		t.Fatalf("file : %v", err)
	}
	if stub.queueStatus == nil || *stub.queueStatus != status {
		t.Fatalf("filtre non transmis : %v", stub.queueStatus)
	}
}

func TestListModerationQueueRejectsUnknownStatus(t *testing.T) {
	stub := &contributionsRepositoryStub{}
	if _, err := newContribTestService(stub).ListModerationQueue(context.Background(), domain.ModerationFilter{Status: "draft", Limit: 20, Offset: 0}); !errors.Is(err, ErrInvalidModeration) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrInvalidModeration, err)
	}
}

func TestListModerationQueueValidatesPagination(t *testing.T) {
	stub := &contributionsRepositoryStub{}
	if _, err := newContribTestService(stub).ListModerationQueue(context.Background(), domain.ModerationFilter{Limit: 0, Offset: 0}); !errors.Is(err, ErrInvalidPagination) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrInvalidPagination, err)
	}
}

func TestApproveContribution(t *testing.T) {
	stub := &contributionsRepositoryStub{productDetail: productContributionDetail()}
	detail, err := newContribTestService(stub).ReviewContribution(context.Background(), contributorUserID, contributionID, approveInput(""))
	if err != nil {
		t.Fatalf("approbation : %v", err)
	}
	if detail.Status != domain.StatusApproved {
		t.Fatalf("statut %s, attendu approved", detail.Status)
	}
	if stub.review.ID != contributionID || stub.review.Type != "product_price" || stub.review.NewStatus != domain.StatusApproved {
		t.Fatalf("revue inattendue : %+v", stub.review)
	}
	if stub.review.ModeratorID != contributorUserID {
		t.Fatalf("modérateur %s, attendu %s", stub.review.ModeratorID, contributorUserID)
	}
}

func TestApproveFuelContribution(t *testing.T) {
	stub := &contributionsRepositoryStub{fuelDetail: fuelContributionDetail()}
	detail, err := newContribTestService(stub).ReviewContribution(context.Background(), contributorUserID, contributionID, approveInput(""))
	if err != nil {
		t.Fatalf("approbation carburant : %v", err)
	}
	if detail.Status != domain.StatusApproved || stub.review.Type != "fuel_price" {
		t.Fatalf("revue inattendue : %+v / %+v", detail, stub.review)
	}
}

func TestRejectContributionRequiresNote(t *testing.T) {
	stub := &contributionsRepositoryStub{productDetail: productContributionDetail()}
	if _, err := newContribTestService(stub).ReviewContribution(context.Background(), contributorUserID, contributionID, domain.ContributionReviewInput{Decision: domain.StatusRejected}); !errors.Is(err, ErrInvalidModeration) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrInvalidModeration, err)
	}
	blankNote := "   "
	if _, err := newContribTestService(stub).ReviewContribution(context.Background(), contributorUserID, contributionID, domain.ContributionReviewInput{Decision: domain.StatusRejected, Note: &blankNote}); !errors.Is(err, ErrInvalidModeration) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrInvalidModeration, err)
	}
}

func TestRejectContribution(t *testing.T) {
	stub := &contributionsRepositoryStub{productDetail: productContributionDetail()}
	detail, err := newContribTestService(stub).ReviewContribution(context.Background(), contributorUserID, contributionID, rejectInput("Prix hors fourchette"))
	if err != nil {
		t.Fatalf("rejet : %v", err)
	}
	if detail.Status != domain.StatusRejected {
		t.Fatalf("statut %s, attendu rejected", detail.Status)
	}
	if stub.review.Note == nil || *stub.review.Note != "Prix hors fourchette" {
		t.Fatalf("note non transmise : %+v", stub.review)
	}
}

func TestReviewRejectsUnknownDecision(t *testing.T) {
	stub := &contributionsRepositoryStub{productDetail: productContributionDetail()}
	if _, err := newContribTestService(stub).ReviewContribution(context.Background(), contributorUserID, contributionID, domain.ContributionReviewInput{Decision: "argue"}); !errors.Is(err, ErrInvalidModeration) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrInvalidModeration, err)
	}
}

func TestReviewRejectsInvalidIDs(t *testing.T) {
	stub := &contributionsRepositoryStub{productDetail: productContributionDetail()}
	if _, err := newContribTestService(stub).ReviewContribution(context.Background(), "pas-un-uuid", contributionID, approveInput("")); !errors.Is(err, ErrInvalidModeration) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrInvalidModeration, err)
	}
	if _, err := newContribTestService(stub).ReviewContribution(context.Background(), contributorUserID, "pas-un-uuid", approveInput("")); !errors.Is(err, ErrInvalidModeration) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrInvalidModeration, err)
	}
}

func TestReviewUnknownContributionIsNotFound(t *testing.T) {
	stub := &contributionsRepositoryStub{}
	detail, err := newContribTestService(stub).ReviewContribution(context.Background(), contributorUserID, contributionID, approveInput(""))
	if !errors.Is(err, domain.ErrNotFound) || detail.Status != "" {
		t.Fatalf("erreur %v, détail %+v", err, detail)
	}
}

func TestReviewAlreadyReviewedContributionIsNotModifiable(t *testing.T) {
	detail := productContributionDetail()
	detail.Contribution.Status = domain.StatusApproved
	stub := &contributionsRepositoryStub{productDetail: detail}
	if _, err := newContribTestService(stub).ReviewContribution(context.Background(), contributorUserID, contributionID, approveInput("")); !errors.Is(err, ErrContributionNotModifiable) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrContributionNotModifiable, err)
	}
}

func TestReviewPropagatesRepositoryError(t *testing.T) {
	stub := &contributionsRepositoryStub{productDetail: productContributionDetail(), reviewError: ErrNoCommunitySource}
	if _, err := newContribTestService(stub).ReviewContribution(context.Background(), contributorUserID, contributionID, approveInput("")); !errors.Is(err, ErrNoCommunitySource) {
		t.Fatalf("erreur attendue %v, obtenu : %v", ErrNoCommunitySource, err)
	}
}
