package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/laurentpoirierfr/ojozone/internal/auth"
	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

const (
	// RefreshTokenTTL est la durée de vie d'un jeton de rafraîchissement.
	RefreshTokenTTL = 30 * 24 * time.Hour
	// MinPasswordLength impose un mot de passe d'au moins 8 caractères.
	MinPasswordLength = 8
	// MaxEmailLength borne l'adresse email.
	MaxEmailLength = 254
)

var (
	// ErrInvalidCredentials indique un email ou mot de passe incorrect.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrEmailAlreadyUsed indique qu'un compte possède déjà cet email.
	ErrEmailAlreadyUsed = errors.New("email already used")
	// ErrInvalidProfile indique un profil ou des champs d'inscription invalides.
	ErrInvalidProfile = errors.New("invalid profile")
)

// Register crée un compte membre et ouvre une session.
func (s *OjoZone) Register(ctx context.Context, input domain.RegisterInput) (domain.AuthResult, error) {
	input.Email = normalizeEmail(input.Email)
	input.Locale = strings.TrimSpace(input.Locale)
	if input.Locale == "" {
		input.Locale = domain.LocaleFrench
	}
	if input.DisplayName != nil {
		name := strings.TrimSpace(*input.DisplayName)
		input.DisplayName = &name
	}
	if !validEmail(input.Email) || !validPassword(input.Password) || !validLocale(input.Locale) || input.DisplayName != nil && len(*input.DisplayName) > 80 {
		return domain.AuthResult{}, ErrInvalidProfile
	}
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return domain.AuthResult{}, err
	}
	user, err := s.repository.CreateUser(ctx, input, passwordHash)
	if errors.Is(err, domain.ErrConflict) {
		return domain.AuthResult{}, ErrEmailAlreadyUsed
	}
	if err != nil {
		return domain.AuthResult{}, err
	}
	return s.openSession(ctx, user)
}

// Login vérifie les identifiants et ouvre une session.
func (s *OjoZone) Login(ctx context.Context, input domain.LoginInput) (domain.AuthResult, error) {
	credentials, err := s.repository.GetUserByEmail(ctx, input.Email)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.AuthResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return domain.AuthResult{}, err
	}
	if credentials.PasswordHash == "" || auth.CheckPassword(credentials.PasswordHash, input.Password) != nil {
		return domain.AuthResult{}, ErrInvalidCredentials
	}
	return s.openSession(ctx, credentials.User)
}

// Refresh renouvelle les jetons par rotation du jeton de rafraîchissement.
func (s *OjoZone) Refresh(ctx context.Context, input domain.RefreshInput) (domain.AuthResult, error) {
	raw := strings.TrimSpace(input.RefreshToken)
	if raw == "" {
		return domain.AuthResult{}, domain.ErrUnauthorized
	}
	session, err := s.repository.GetSessionByRefreshHash(ctx, s.auth.HashRefreshToken(raw))
	if errors.Is(err, domain.ErrNotFound) {
		return domain.AuthResult{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.AuthResult{}, err
	}
	now := time.Now()
	if session.RevokedAt != nil || session.ExpiresAt.Before(now) {
		_ = s.repository.RevokeSession(ctx, session.ID)
		return domain.AuthResult{}, domain.ErrUnauthorized
	}
	user, err := s.repository.GetUserByID(ctx, session.UserID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.AuthResult{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.AuthResult{}, err
	}
	newRaw, newHash, err := s.auth.NewRefreshToken()
	if err != nil {
		return domain.AuthResult{}, err
	}
	updated, err := s.repository.UpdateSessionRefresh(ctx, session.ID, newHash, now.Add(RefreshTokenTTL))
	if err != nil {
		return domain.AuthResult{}, err
	}
	access, err := s.auth.IssueAccessToken(user.ID, user.Role, updated.ID)
	if err != nil {
		return domain.AuthResult{}, err
	}
	return domain.AuthResult{User: user, Tokens: tokenPair(access, newRaw, s.auth.AccessTTL())}, nil
}

// Logout révoque la session liée au jeton de rafraîchissement.
func (s *OjoZone) Logout(ctx context.Context, input domain.RefreshInput) error {
	raw := strings.TrimSpace(input.RefreshToken)
	if raw == "" {
		return nil
	}
	session, err := s.repository.GetSessionByRefreshHash(ctx, s.auth.HashRefreshToken(raw))
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.repository.RevokeSession(ctx, session.ID)
}

// GetMe retourne le profil de l'utilisateur courant.
func (s *OjoZone) GetMe(ctx context.Context, userID string) (domain.User, error) {
	if !validUUID(userID) {
		return domain.User{}, ErrInvalidID
	}
	user, err := s.repository.GetUserByID(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, domain.ErrUnauthorized
	}
	return user, err
}

// UpdateMe fusionne les modifications du profil courant avec les valeurs existantes.
func (s *OjoZone) UpdateMe(ctx context.Context, userID string, input domain.UpdateProfileInput) (domain.User, error) {
	if !validUUID(userID) {
		return domain.User{}, ErrInvalidID
	}
	if input.DisplayName != nil && len(*input.DisplayName) > 80 {
		return domain.User{}, ErrInvalidProfile
	}
	if input.Locale != nil && !validLocale(*input.Locale) {
		return domain.User{}, ErrInvalidProfile
	}
	current, err := s.repository.GetUserByID(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.User{}, err
	}
	name := current.DisplayName
	if input.DisplayName != nil {
		name = input.DisplayName
	}
	locale := current.Locale
	if input.Locale != nil {
		locale = *input.Locale
	}
	return s.repository.UpdateUserProfile(ctx, userID, domain.UpdateProfileInput{DisplayName: name, Locale: &locale})
}

// DeleteMe révoque les sessions puis anonymise le compte.
func (s *OjoZone) DeleteMe(ctx context.Context, userID string) error {
	if !validUUID(userID) {
		return ErrInvalidID
	}
	if _, err := s.repository.AnonymizeUser(ctx, userID); errors.Is(err, domain.ErrNotFound) {
		return domain.ErrUnauthorized
	} else if err != nil {
		return err
	}
	return s.repository.RevokeAllSessionsForUser(ctx, userID)
}

// ListMyContributions retourne les observations soumises par l'utilisateur courant.
func (s *OjoZone) ListMyContributions(ctx context.Context, userID string, page domain.Pagination) ([]domain.Contribution, error) {
	if !validUUID(userID) {
		return nil, ErrInvalidID
	}
	if err := validatePagination(page.Limit, page.Offset); err != nil {
		return nil, err
	}
	return s.repository.ListMyContributions(ctx, userID, page)
}

// openSession émet un jeton d'accès et persiste un jeton de rafraîchissement.
func (s *OjoZone) openSession(ctx context.Context, user domain.User) (domain.AuthResult, error) {
	raw, hash, err := s.auth.NewRefreshToken()
	if err != nil {
		return domain.AuthResult{}, err
	}
	session, err := s.repository.CreateSession(ctx, domain.Session{
		UserID: user.ID, RefreshTokenHash: hash, ExpiresAt: time.Now().Add(RefreshTokenTTL),
	})
	if err != nil {
		return domain.AuthResult{}, err
	}
	access, err := s.auth.IssueAccessToken(user.ID, user.Role, session.ID)
	if err != nil {
		return domain.AuthResult{}, err
	}
	return domain.AuthResult{User: user, Tokens: tokenPair(access, raw, s.auth.AccessTTL())}, nil
}

func tokenPair(access, refresh string, ttl time.Duration) domain.TokenPair {
	return domain.TokenPair{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", ExpiresIn: int64(ttl.Seconds())}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	if len(email) == 0 || len(email) > MaxEmailLength || strings.ContainsAny(email, " \t\r\n") {
		return false
	}
	at := strings.IndexByte(email, '@')
	return at > 0 && at < len(email)-1 && strings.Contains(email[at+1:], ".")
}

func validPassword(password string) bool {
	return len(password) >= MinPasswordLength && len(password) <= 128
}

func validLocale(locale string) bool {
	return locale == domain.LocaleFrench || locale == domain.LocaleEnglish
}
