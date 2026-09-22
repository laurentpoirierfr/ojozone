package domain

import "time"

// Statuts de modération d'une contribution ou d'une observation.
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusFlagged  = "flagged"
)

// ProductContributionInput propose un prix produit (statut pending).
type ProductContributionInput struct {
	ProductID   string    `json:"product_id" example:"2bded1b3-5cb0-4a43-aee2-6b41bfbcb8dd" format:"uuid"`
	LocationID  string    `json:"location_id" example:"1b368d0e-61a1-4af7-8523-6f447330866d" format:"uuid"`
	Amount      string    `json:"amount" example:"2.35"`
	Currency    string    `json:"currency" example:"EUR" minLength:"3" maxLength:"3"`
	Quantity    string    `json:"quantity" example:"1.000"`
	UnitCode    string    `json:"unit_code" example:"L" maxLength:"16"`
	IsPromotion bool      `json:"is_promotion"`
	ObservedAt  time.Time `json:"observed_at" format:"date-time"`
}

// FuelContributionInput propose un prix carburant au litre (statut pending).
type FuelContributionInput struct {
	FuelTypeID     string    `json:"fuel_type_id" example:"b25fbc0e-9d7a-4d2e-bf5f-7d2a1f0a3c92" format:"uuid"`
	LocationID     string    `json:"location_id" example:"1b368d0e-61a1-4af7-8523-6f447330866d" format:"uuid"`
	AmountPerLitre string    `json:"amount_per_litre" example:"1.849"`
	Currency       string    `json:"currency" example:"EUR" minLength:"3" maxLength:"3"`
	ObservedAt     time.Time `json:"observed_at" format:"date-time"`
}

// ProductContributionSubmit est la forme persistée d'un prix proposé.
type ProductContributionSubmit struct {
	ProductContributionInput
	ContributorID    string
	SourceID         string
	NormalizedAmount string
}

// FuelContributionSubmit est la forme persistée d'un carburant proposé.
type FuelContributionSubmit struct {
	FuelContributionInput
	ContributorID string
	SourceID      string
}

// ContributionCheck nomme un contrôle passé avant acceptation.
type ContributionCheck struct {
	Code string `json:"code" example:"format_valid"`
}

// ContributionResult expose le statut d'une soumission.
type ContributionResult struct {
	ID     string              `json:"id" example:"edfb822d-d78a-481e-9945-850e337a33fd"`
	Status string              `json:"status" example:"pending"`
	Checks []ContributionCheck `json:"checks"`
}

// ContributionPatch corrige une contribution encore modifiable (pending).
type ContributionPatch struct {
	Amount         *string    `json:"amount,omitempty" example:"2.45"`
	Currency       *string    `json:"currency,omitempty" example:"EUR" minLength:"3" maxLength:"3"`
	Quantity       *string    `json:"quantity,omitempty" example:"1.000"`
	UnitCode       *string    `json:"unit_code,omitempty" example:"L" maxLength:"16"`
	IsPromotion    *bool      `json:"is_promotion,omitempty"`
	AmountPerLitre *string    `json:"amount_per_litre,omitempty" example:"1.899"`
	ObservedAt     *time.Time `json:"observed_at,omitempty" format:"date-time"`
}

// ContributionDetail est une contribution avec son auteur, pour les modérateurs et le propriétaire.
type ContributionDetail struct {
	Contribution
	ContributorID string `json:"contributor_id" format:"uuid"`
}
