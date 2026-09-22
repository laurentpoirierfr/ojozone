package repository

import (
	"context"
	"errors"
	"net/netip"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
	"github.com/laurentpoirierfr/ojozone/internal/store/db"
)

func (r *PostgreSQL) CreateUser(ctx context.Context, input domain.RegisterInput, passwordHash string) (domain.User, error) {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: passwordHash,
		DisplayName:  input.DisplayName,
		Locale:       input.Locale,
	})
	if err != nil {
		return domain.User{}, mapWriteError(err)
	}
	return userFromCreateRow(row), nil
}

func (r *PostgreSQL) GetUserByEmail(ctx context.Context, email string) (domain.UserWithPassword, error) {
	row, err := r.queries.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.UserWithPassword{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.UserWithPassword{}, err
	}
	return domain.UserWithPassword{
		User:         userFromTable(row),
		PasswordHash: fromNullableString(row.PasswordHash),
	}, nil
}

func (r *PostgreSQL) GetUserByID(ctx context.Context, id string) (domain.User, error) {
	row, err := r.queries.GetUserById(ctx, postgresUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	return userFromGetRow(row), nil
}

func (r *PostgreSQL) UpdateUserProfile(ctx context.Context, id string, input domain.UpdateProfileInput) (domain.User, error) {
	var name *string
	if input.DisplayName != nil {
		trimmed := strings.TrimSpace(*input.DisplayName)
		if trimmed == "" {
			name = nil
		} else {
			name = &trimmed
		}
	}
	locale := "fr"
	if input.Locale != nil {
		locale = *input.Locale
	}
	row, err := r.queries.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID: postgresUUID(id), DisplayName: name, Locale: locale,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	return userFromUpdateRow(row), nil
}

func (r *PostgreSQL) AnonymizeUser(ctx context.Context, id string) (domain.User, error) {
	row, err := r.queries.AnonymizeUser(ctx, postgresUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	return userFromAnonymizeRow(row), nil
}

func (r *PostgreSQL) CreateSession(ctx context.Context, session domain.Session) (domain.Session, error) {
	row, err := r.queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:           postgresUUID(session.UserID),
		RefreshTokenHash: session.RefreshTokenHash,
		ExpiresAt:        pgtype.Timestamptz{Time: session.ExpiresAt, Valid: true},
		UserAgent:        session.UserAgent,
		IpAddress:        parseOptionalAddr(session.IPAddress),
	})
	if err != nil {
		return domain.Session{}, err
	}
	return sessionFromRow(row), nil
}

func (r *PostgreSQL) GetSessionByRefreshHash(ctx context.Context, hash string) (domain.Session, error) {
	row, err := r.queries.GetSessionByRefreshHash(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Session{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Session{}, err
	}
	return sessionFromRow(row), nil
}

func (r *PostgreSQL) UpdateSessionRefresh(ctx context.Context, id, refreshTokenHash string, expiresAt time.Time) (domain.Session, error) {
	row, err := r.queries.UpdateSessionRefresh(ctx, db.UpdateSessionRefreshParams{
		ID: postgresUUID(id), RefreshTokenHash: refreshTokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Session{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.Session{}, err
	}
	return sessionFromRow(row), nil
}

func (r *PostgreSQL) RevokeSession(ctx context.Context, id string) error {
	_, err := r.queries.RevokeSession(ctx, postgresUUID(id))
	return err
}

func (r *PostgreSQL) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	_, err := r.queries.RevokeAllSessionsForUser(ctx, postgresUUID(userID))
	return err
}

func (r *PostgreSQL) ListMyContributions(ctx context.Context, userID string, page domain.Pagination) ([]domain.Contribution, error) {
	productRows, err := r.queries.ListProductPriceContributions(ctx, db.ListProductPriceContributionsParams{
		ContributorID: postgresUUID(userID), PageSize: page.Limit, PageOffset: page.Offset,
	})
	if err != nil {
		return nil, err
	}
	fuelRows, err := r.queries.ListFuelPriceContributions(ctx, db.ListFuelPriceContributionsParams{
		ContributorID: postgresUUID(userID), PageSize: page.Limit, PageOffset: page.Offset,
	})
	if err != nil {
		return nil, err
	}
	contributions := make([]domain.Contribution, 0, len(productRows)+len(fuelRows))
	for _, row := range productRows {
		contributions = append(contributions, productContribution(row))
	}
	for _, row := range fuelRows {
		contributions = append(contributions, fuelContribution(row))
	}
	return contributions, nil
}

func fromNullableString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func parseOptionalAddr(value *string) *netip.Addr {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	parsed, err := netip.ParseAddr(*value)
	if err != nil {
		return nil
	}
	return &parsed
}

func sessionFromRow(row db.Session) domain.Session {
	return domain.Session{
		ID: uuidString(row.ID), UserID: uuidString(row.UserID), RefreshTokenHash: row.RefreshTokenHash,
		ExpiresAt: row.ExpiresAt.Time, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
		RevokedAt: optionalTime(row.RevokedAt), UserAgent: row.UserAgent,
		IPAddress: optionalAddrString(row.IpAddress),
	}
}

func optionalAddrString(value *netip.Addr) *string {
	if value == nil {
		return nil
	}
	text := value.String()
	return &text
}

func userFromTable(row db.User) domain.User {
	return domain.User{
		ID: uuidString(row.ID), Email: row.Email, DisplayName: row.DisplayName,
		Role: string(row.Role), Locale: row.Locale, ReputationScore: row.ReputationScore,
		EmailVerifiedAt: optionalTime(row.EmailVerifiedAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func userFromCreateRow(row db.CreateUserRow) domain.User {
	return domain.User{
		ID: uuidString(row.ID), Email: row.Email, DisplayName: row.DisplayName,
		Role: string(row.Role), Locale: row.Locale, ReputationScore: row.ReputationScore,
		EmailVerifiedAt: optionalTime(row.EmailVerifiedAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func userFromGetRow(row db.GetUserByIdRow) domain.User {
	return domain.User{
		ID: uuidString(row.ID), Email: row.Email, DisplayName: row.DisplayName,
		Role: string(row.Role), Locale: row.Locale, ReputationScore: row.ReputationScore,
		EmailVerifiedAt: optionalTime(row.EmailVerifiedAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func userFromUpdateRow(row db.UpdateUserProfileRow) domain.User {
	return domain.User{
		ID: uuidString(row.ID), Email: row.Email, DisplayName: row.DisplayName,
		Role: string(row.Role), Locale: row.Locale, ReputationScore: row.ReputationScore,
		EmailVerifiedAt: optionalTime(row.EmailVerifiedAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func userFromAnonymizeRow(row db.AnonymizeUserRow) domain.User {
	return domain.User{
		ID: uuidString(row.ID), Email: row.Email, DisplayName: row.DisplayName,
		Role: string(row.Role), Locale: row.Locale, ReputationScore: row.ReputationScore,
		EmailVerifiedAt: optionalTime(row.EmailVerifiedAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func productContribution(row db.ListProductPriceContributionsRow) domain.Contribution {
	quantity := numericString(row.Quantity)
	fuelUnit := strings.TrimSpace(row.UnitCode)
	return domain.Contribution{
		ID: uuidString(row.ID), Type: "product_price", Status: string(row.Status),
		ObservedAt: row.ObservedAt.Time, CreatedAt: row.CreatedAt.Time,
		Currency: strings.TrimSpace(row.Currency), Amount: numericString(row.Amount),
		Quantity: &quantity, UnitCode: &fuelUnit,
		ConfidenceScore: nullableNumericString(row.ConfidenceScore), Subject: row.ProductName,
		Location: domain.EntityRef{ID: uuidString(row.LocationID), Name: row.LocationName},
		GeoArea:  domain.EntityRef{ID: uuidString(row.GeoAreaID), Name: row.GeoAreaName},
		Source:   domain.SourceRef{ID: uuidString(row.SourceID), Name: row.SourceName, Kind: string(row.SourceKind)},
	}
}

func fuelContribution(row db.ListFuelPriceContributionsRow) domain.Contribution {
	return domain.Contribution{
		ID: uuidString(row.ID), Type: "fuel_price", Status: string(row.Status),
		ObservedAt: row.ObservedAt.Time, CreatedAt: row.CreatedAt.Time,
		Currency: strings.TrimSpace(row.Currency), Amount: numericString(row.AmountPerLitre),
		ConfidenceScore: nullableNumericString(row.ConfidenceScore), Subject: row.FuelTypeCode,
		Location: domain.EntityRef{ID: uuidString(row.LocationID), Name: row.LocationName},
		GeoArea:  domain.EntityRef{ID: uuidString(row.GeoAreaID), Name: row.GeoAreaName},
		Source:   domain.SourceRef{ID: uuidString(row.SourceID), Name: row.SourceName, Kind: string(row.SourceKind)},
	}
}
