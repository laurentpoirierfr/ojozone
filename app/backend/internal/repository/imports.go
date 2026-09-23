package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
	"github.com/laurentpoirierfr/ojozone/internal/store/db"
)

// CreateImport enregistre atomiquement un import et ses lignes.
func (r *PostgreSQL) CreateImport(ctx context.Context, input domain.ImportCreate, createdBy string) (domain.Import, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Import{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	transactions := r.queries.WithTx(tx)
	row, err := transactions.CreateImport(ctx, db.CreateImportParams{
		ResourceType: input.ResourceType,
		CreatedBy:    postgresUUID(createdBy),
		LineCount:    int32(len(input.Rows)),
	})
	if err != nil {
		return domain.Import{}, err
	}
	for i, payload := range input.Rows {
		if err := transactions.AddImportRow(ctx, db.AddImportRowParams{
			ImportID:   row.ID,
			LineNumber: int32(i + 1),
			Payload:    payload,
		}); err != nil {
			return domain.Import{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Import{}, err
	}
	return importFromRow(row), nil
}

// GetImport retourne un import par identifiant.
func (r *PostgreSQL) GetImport(ctx context.Context, id string) (domain.Import, error) {
	row, err := r.queries.GetImport(ctx, postgresUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Import{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Import{}, err
	}
	return importFromRow(row), nil
}

// ListImports retourne les imports du plus récent au plus ancien.
func (r *PostgreSQL) ListImports(ctx context.Context, pagination domain.Pagination) ([]domain.Import, error) {
	rows, err := r.queries.ListImports(ctx, db.ListImportsParams{
		PageSize: pagination.Limit, PageOffset: pagination.Offset,
	})
	if err != nil {
		return nil, err
	}
	imports := make([]domain.Import, 0, len(rows))
	for _, row := range rows {
		imports = append(imports, importFromRow(row))
	}
	return imports, nil
}

// ListImportRows retourne les lignes d'un import dans l'ordre des numéros.
func (r *PostgreSQL) ListImportRows(ctx context.Context, importID string) ([]domain.ImportRow, error) {
	rows, err := r.queries.ListImportRows(ctx, postgresUUID(importID))
	if err != nil {
		return nil, err
	}
	items := make([]domain.ImportRow, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.ImportRow{
			ID:         uuidString(row.ID),
			LineNumber: row.LineNumber,
			Payload:    row.Payload,
			Valid:      row.Valid != nil && *row.Valid,
			Error:      row.Error,
		})
	}
	return items, nil
}

// SetImportRowStatus mémorise le résultat d'une ligne après validation.
func (r *PostgreSQL) SetImportRowStatus(ctx context.Context, importID string, lineNumber int32, valid bool, message *string) error {
	return r.queries.SetImportRowStatus(ctx, db.SetImportRowStatusParams{
		ImportID: postgresUUID(importID), LineNumber: lineNumber, Valid: valid, Error: message,
	})
}

// MarkImportValidated passe un import au statut validé avec son rapport.
func (r *PostgreSQL) MarkImportValidated(ctx context.Context, id string, validCount, invalidCount int32, report []byte) error {
	return r.queries.MarkImportValidated(ctx, db.MarkImportValidatedParams{
		ID: postgresUUID(id), ValidCount: validCount, InvalidCount: invalidCount, Report: report,
	})
}

// MarkImportPublished passe un import au statut publié.
func (r *PostgreSQL) MarkImportPublished(ctx context.Context, id string, report []byte) error {
	return r.queries.MarkImportPublished(ctx, db.MarkImportPublishedParams{
		ID: postgresUUID(id), Report: report,
	})
}

func importFromRow(row db.Import) domain.Import {
	return domain.Import{
		ID:           uuidString(row.ID),
		ResourceType: row.ResourceType,
		Status:       row.Status,
		CreatedBy:    optionalUUID(row.CreatedBy),
		LineCount:    row.LineCount,
		ValidCount:   row.ValidCount,
		InvalidCount: row.InvalidCount,
		Report:       validJSON(row.Report, "[]"),
		CreatedAt:    row.CreatedAt.Time,
		ValidatedAt:  optionalTime(row.ValidatedAt),
		PublishedAt:  optionalTime(row.PublishedAt),
	}
}
