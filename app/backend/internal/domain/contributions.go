package domain

import (
	"errors"
	"time"
)

// ErrContributionNotModifiable indique qu'une contribution a déjà été traitée.
var ErrContributionNotModifiable = errors.New("contribution no longer modifiable")

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

// ModerationFilter filtre la file de modération.
type ModerationFilter struct {
	Status string
	Limit  int32
	Offset int32
}

// ModerationQueueItem est une entrée de la file de modération.
type ModerationQueueItem struct {
	ID               string    `json:"id" format:"uuid"`
	Type             string    `json:"type"`
	Status           string    `json:"status"`
	Subject          string    `json:"subject"`
	ContributorID    *string   `json:"contributor_id,omitempty" format:"uuid"`
	ContributorEmail *string   `json:"contributor_email,omitempty"`
	Amount           string    `json:"amount"`
	Currency         string    `json:"currency"`
	Quantity         *string   `json:"quantity,omitempty"`
	UnitCode         *string   `json:"unit_code,omitempty"`
	IsPromotion      *bool     `json:"is_promotion,omitempty"`
	Location         EntityRef `json:"location"`
	GeoArea          EntityRef `json:"geo_area"`
	Source           SourceRef `json:"source"`
	ObservedAt       time.Time `json:"observed_at"`
	ConfidenceScore  *string   `json:"confidence_score,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// ContributionReviewInput est la décision d'un modérateur sur une contribution.
type ContributionReviewInput struct {
	Decision   string  `json:"decision" example:"approved" enums:"approved,rejected"`
	Note       *string `json:"note,omitempty" example:"Prix cohérent avec les relevés voisins"`
	ReasonCode *string `json:"reason_code,omitempty" example:"outlier"`
}

// ContributionReview est la décision prête à être appliquée en base.
type ContributionReview struct {
	ID          string
	Type        string
	ModeratorID string
	NewStatus   string
	ReasonCode  *string
	Note        *string
}
