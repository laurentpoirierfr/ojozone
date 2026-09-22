package service

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

// SubmitProductContribution propose un prix produit avec le statut pending.
func (s *OjoZone) SubmitProductContribution(ctx context.Context, userID string, input domain.ProductContributionInput) (domain.ContributionResult, error) {
	if err := normalizeAndValidateProductContribution(&input); err != nil {
		return domain.ContributionResult{}, err
	}
	sourceID, err := s.communitySourceID(ctx)
	if err != nil {
		return domain.ContributionResult{}, err
	}
	id, status, err := s.repository.CreateProductContribution(ctx, domain.ProductContributionSubmit{
		ProductContributionInput: input,
		ContributorID:            userID,
		SourceID:                 sourceID,
		NormalizedAmount:         normalizedAmount(input.Amount, input.Quantity),
	})
	if err != nil {
		return domain.ContributionResult{}, err
	}
	return domain.ContributionResult{
		ID:     id,
		Status: status,
		Checks: []domain.ContributionCheck{{Code: "format_valid"}},
	}, nil
}

// SubmitFuelContribution propose un prix carburant avec le statut pending.
func (s *OjoZone) SubmitFuelContribution(ctx context.Context, userID string, input domain.FuelContributionInput) (domain.ContributionResult, error) {
	if err := normalizeAndValidateFuelContribution(&input); err != nil {
		return domain.ContributionResult{}, err
	}
	sourceID, err := s.communitySourceID(ctx)
	if err != nil {
		return domain.ContributionResult{}, err
	}
	id, status, err := s.repository.CreateFuelContribution(ctx, domain.FuelContributionSubmit{
		FuelContributionInput: input,
		ContributorID:         userID,
		SourceID:              sourceID,
	})
	if err != nil {
		return domain.ContributionResult{}, err
	}
	return domain.ContributionResult{
		ID:     id,
		Status: status,
		Checks: []domain.ContributionCheck{{Code: "format_valid"}},
	}, nil
}

// GetContribution expose une contribution au propriétaire, aux modérateurs et aux administrateurs.
func (s *OjoZone) GetContribution(ctx context.Context, userID, role, id string) (domain.ContributionDetail, error) {
	detail, _, err := s.fetchContribution(ctx, id)
	if err != nil {
		return domain.ContributionDetail{}, err
	}
	if !canManageContribution(role, detail.ContributorID, userID) {
		return domain.ContributionDetail{}, domain.ErrForbidden
	}
	return detail, nil
}

// UpdateContribution corrige une contribution pending appartenant à l'utilisateur courant.
func (s *OjoZone) UpdateContribution(ctx context.Context, userID, id string, patch domain.ContributionPatch) (domain.ContributionDetail, error) {
	detail, contributionType, err := s.fetchContribution(ctx, id)
	if err != nil {
		return domain.ContributionDetail{}, err
	}
	if detail.ContributorID != userID {
		return domain.ContributionDetail{}, domain.ErrForbidden
	}
	if detail.Status != domain.StatusPending {
		return domain.ContributionDetail{}, ErrContributionNotModifiable
	}
	switch contributionType {
	case "product_price":
		submit := domain.ProductContributionSubmit{
			ProductContributionInput: mergeProductContribution(detail, patch),
			ContributorID:            userID,
			NormalizedAmount:         "",
		}
		if err := validateProductContributionFields(&submit.ProductContributionInput); err != nil {
			return domain.ContributionDetail{}, err
		}
		submit.NormalizedAmount = normalizedAmount(submit.Amount, submit.Quantity)
		if err := s.repository.UpdateProductContribution(ctx, id, userID, submit); err != nil {
			return domain.ContributionDetail{}, err
		}
	case "fuel_price":
		submit := domain.FuelContributionSubmit{
			FuelContributionInput: mergeFuelContribution(detail, patch),
			ContributorID:         userID,
		}
		if err := validateFuelContributionFields(&submit.FuelContributionInput); err != nil {
			return domain.ContributionDetail{}, err
		}
		if err := s.repository.UpdateFuelContribution(ctx, id, userID, submit); err != nil {
			return domain.ContributionDetail{}, err
		}
	default:
		return domain.ContributionDetail{}, domain.ErrNotFound
	}
	return s.fetchContributionAndCheck(ctx, userID, id)
}

// DeleteContribution retire une contribution pending appartenant à l'utilisateur courant.
func (s *OjoZone) DeleteContribution(ctx context.Context, userID, id string) error {
	detail, contributionType, err := s.fetchContribution(ctx, id)
	if err != nil {
		return err
	}
	if detail.ContributorID != userID {
		return domain.ErrForbidden
	}
	if detail.Status != domain.StatusPending {
		return ErrContributionNotModifiable
	}
	if contributionType == "product_price" {
		return s.repository.DeleteProductContribution(ctx, id, userID)
	}
	return s.repository.DeleteFuelContribution(ctx, id, userID)
}

func (s *OjoZone) communitySourceID(ctx context.Context) (string, error) {
	sourceID, err := s.repository.FindCommunitySourceID(ctx)
	if errors.Is(err, domain.ErrNotFound) {
		return "", ErrNoCommunitySource
	}
	if err != nil {
		return "", err
	}
	return sourceID, nil
}

// fetchContribution résout une contribution produit ou carburant et retourne son type.
func (s *OjoZone) fetchContribution(ctx context.Context, id string) (domain.ContributionDetail, string, error) {
	if !validUUID(id) {
		return domain.ContributionDetail{}, "", ErrInvalidID
	}
	detail, err := s.repository.GetProductContribution(ctx, id)
	if err == nil {
		return detail, "product_price", nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.ContributionDetail{}, "", err
	}
	detail, err = s.repository.GetFuelContribution(ctx, id)
	if err != nil {
		return domain.ContributionDetail{}, "", err
	}
	return detail, "fuel_price", nil
}

func (s *OjoZone) fetchContributionAndCheck(ctx context.Context, userID, id string) (domain.ContributionDetail, error) {
	detail, _, err := s.fetchContribution(ctx, id)
	if err != nil {
		return domain.ContributionDetail{}, err
	}
	if detail.ContributorID != userID {
		return domain.ContributionDetail{}, domain.ErrNotFound
	}
	return detail, nil
}

func canManageContribution(role, contributorID, userID string) bool {
	if role == domain.RoleModerator || role == domain.RoleAdmin {
		return true
	}
	return contributorID == userID
}

func normalizedAmount(amount, quantity string) string {
	amountRat, ok := new(big.Rat).SetString(amount)
	if !ok {
		return ""
	}
	quantityRat, ok := new(big.Rat).SetString(quantity)
	if !ok || quantityRat.Sign() == 0 {
		return ""
	}
	quotient := new(big.Rat).Quo(amountRat, quantityRat)
	return quotient.FloatString(4)
}

func normalizeAndValidateProductContribution(input *domain.ProductContributionInput) error {
	if !validUUID(input.ProductID) || !validUUID(input.LocationID) {
		return ErrInvalidContribution
	}
	return validateProductContributionFields(input)
}

func validateProductContributionFields(input *domain.ProductContributionInput) error {
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.UnitCode = strings.TrimSpace(input.UnitCode)
	if len(input.Currency) != 3 || input.UnitCode == "" || len(input.UnitCode) > 16 || input.ObservedAt.IsZero() {
		return ErrInvalidContribution
	}
	if !validNonNegativeDecimal(input.Amount) || !validPositiveDecimal(input.Quantity) {
		return ErrInvalidContribution
	}
	return nil
}

func normalizeAndValidateFuelContribution(input *domain.FuelContributionInput) error {
	if !validUUID(input.FuelTypeID) || !validUUID(input.LocationID) {
		return ErrInvalidContribution
	}
	return validateFuelContributionFields(input)
}

func validateFuelContributionFields(input *domain.FuelContributionInput) error {
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	if len(input.Currency) != 3 || input.ObservedAt.IsZero() {
		return ErrInvalidContribution
	}
	if !validPositiveDecimal(input.AmountPerLitre) {
		return ErrInvalidContribution
	}
	return nil
}

func mergeProductContribution(detail domain.ContributionDetail, patch domain.ContributionPatch) domain.ProductContributionInput {
	input := domain.ProductContributionInput{
		Amount:      detail.Amount,
		Currency:    detail.Currency,
		Quantity:    fromStringPtr(detail.Quantity),
		UnitCode:    fromStringPtr(detail.UnitCode),
		IsPromotion: boolFromPtr(detail.IsPromotion),
		ObservedAt:  detail.ObservedAt,
	}
	if patch.Amount != nil {
		input.Amount = *patch.Amount
	}
	if patch.Currency != nil {
		input.Currency = *patch.Currency
	}
	if patch.Quantity != nil {
		input.Quantity = *patch.Quantity
	}
	if patch.UnitCode != nil {
		input.UnitCode = *patch.UnitCode
	}
	if patch.IsPromotion != nil {
		input.IsPromotion = *patch.IsPromotion
	}
	if patch.ObservedAt != nil {
		input.ObservedAt = *patch.ObservedAt
	}
	return input
}

func mergeFuelContribution(detail domain.ContributionDetail, patch domain.ContributionPatch) domain.FuelContributionInput {
	input := domain.FuelContributionInput{
		AmountPerLitre: detail.Amount,
		Currency:       detail.Currency,
		ObservedAt:     detail.ObservedAt,
	}
	if patch.AmountPerLitre != nil {
		input.AmountPerLitre = *patch.AmountPerLitre
	}
	if patch.Currency != nil {
		input.Currency = *patch.Currency
	}
	if patch.ObservedAt != nil {
		input.ObservedAt = *patch.ObservedAt
	}
	return input
}

func fromStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func boolFromPtr(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}
