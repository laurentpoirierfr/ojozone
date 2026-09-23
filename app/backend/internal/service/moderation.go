package service

import (
	"context"
	"strings"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

// ListModerationQueue retourne les contributions à contrôler, filtrées par statut.
func (s *OjoZone) ListModerationQueue(ctx context.Context, filter domain.ModerationFilter) ([]domain.ModerationQueueItem, error) {
	filter.Status = strings.TrimSpace(filter.Status)
	if filter.Status != "" && !validStatus(filter.Status) {
		return nil, ErrInvalidModeration
	}
	if err := validatePagination(filter.Limit, filter.Offset); err != nil {
		return nil, err
	}
	var status *string
	if filter.Status != "" {
		status = &filter.Status
	}
	return s.repository.ListModerationQueue(ctx, status, domain.Pagination{Limit: filter.Limit, Offset: filter.Offset})
}

// ReviewContribution applique la décision d'un modérateur sur une contribution.
// Une contribution n'est modifiable que depuis les statuts pending ou flagged ;
// un rejet exige une note. La décision est auditée en append-only.
func (s *OjoZone) ReviewContribution(ctx context.Context, moderatorID, id string, input domain.ContributionReviewInput) (domain.ContributionDetail, error) {
	if !validUUID(id) || !validUUID(moderatorID) {
		return domain.ContributionDetail{}, ErrInvalidModeration
	}
	decision := strings.TrimSpace(input.Decision)
	switch decision {
	case domain.StatusApproved:
	case domain.StatusRejected:
		if input.Note == nil || strings.TrimSpace(*input.Note) == "" {
			return domain.ContributionDetail{}, ErrInvalidModeration
		}
	default:
		return domain.ContributionDetail{}, ErrInvalidModeration
	}

	detail, contributionType, err := s.fetchContribution(ctx, id)
	if err != nil {
		return domain.ContributionDetail{}, err
	}
	if detail.Status != domain.StatusPending && detail.Status != domain.StatusFlagged {
		return domain.ContributionDetail{}, ErrContributionNotModifiable
	}

	review := domain.ContributionReview{
		ID:          id,
		Type:        contributionType,
		ModeratorID: moderatorID,
		NewStatus:   decision,
		ReasonCode:  input.ReasonCode,
		Note:        input.Note,
	}
	if err := s.repository.ReviewContribution(ctx, review); err != nil {
		return domain.ContributionDetail{}, err
	}
	revised, _, err := s.fetchContribution(ctx, id)
	return revised, err
}
