package domain

import (
	"errors"
	"time"
)

var (
	// ErrUnauthorized indique qu'un jeton est manquant, invalide ou expiré.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden indique que le rôle ne permet pas l'opération.
	ErrForbidden = errors.New("forbidden")
)

// Rôles utilisateur portés par les jetons d'accès.
const (
	RoleMember    = "member"
	RoleModerator = "moderator"
	RoleAdmin     = "admin"
	RolePartner   = "partner"
)

// Language codes pris en charge par l'interface.
const (
	LocaleFrench  = "fr"
	LocaleEnglish = "en"
)

// User expose un compte sans donnée sensible.
type User struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	DisplayName     *string    `json:"display_name"`
	Role            string     `json:"role"`
	Locale          string     `json:"locale"`
	ReputationScore int32      `json:"reputation_score"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// UserWithPassword contient en plus l'empreinte du mot de passe, destinée uniquement à la vérification.
type UserWithPassword struct {
	User
	PasswordHash string
}

// RegisterInput crée un compte membre.
type RegisterInput struct {
	Email       string  `json:"email" example:"marie@example.com"`
	Password    string  `json:"password" minLength:"8"`
	DisplayName *string `json:"display_name,omitempty" maxLength:"80"`
	Locale      string  `json:"locale" example:"fr" enums:"fr,en"`
}

// LoginInput ouvre une session par email et mot de passe.
type LoginInput struct {
	Email    string `json:"email" example:"marie@example.com"`
	Password string `json:"password"`
}

// RefreshInput renouvelle les jetons à partir d'un jeton de rafraîchissement.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenPair assemble un jeton d'accès court et un jeton de rafraîchissement.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// AuthResult associe le profil et les jetons émis.
type AuthResult struct {
	User   User      `json:"user"`
	Tokens TokenPair `json:"tokens"`
}

// UpdateProfileInput personnalise le profil courant.
type UpdateProfileInput struct {
	DisplayName *string `json:"display_name,omitempty" maxLength:"80"`
	Locale      *string `json:"locale,omitempty" enums:"fr,en"`
}

// Session représente un jeton de rafraîchissement en base.
type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	RevokedAt        *time.Time
	UserAgent        *string
	IPAddress        *string
}

// Contribution est une observation soumise par l'utilisateur courant.
type Contribution struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	ObservedAt      time.Time `json:"observed_at"`
	CreatedAt       time.Time `json:"created_at"`
	Currency        string    `json:"currency"`
	Amount          string    `json:"amount"`
	Quantity        *string   `json:"quantity,omitempty"`
	UnitCode        *string   `json:"unit_code,omitempty"`
	IsPromotion     *bool     `json:"is_promotion,omitempty"`
	ConfidenceScore *string   `json:"confidence_score,omitempty"`
	Subject         string    `json:"subject"`
	Location        EntityRef `json:"location"`
	GeoArea         EntityRef `json:"geo_area"`
	Source          SourceRef `json:"source"`
}
