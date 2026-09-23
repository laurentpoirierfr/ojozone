package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

const MaxImportRows = 10000

var importableResources = map[string]bool{
	ResourceGeoAreas: true, ResourceSources: true, ResourceCategories: true, ResourceUnits: true,
	ResourceMerchants: true, ResourceLocations: true, ResourceFuelTypes: true, ResourceFuelPrices: true,
	ResourceHousing: true, ResourceIncome: true, ResourceProducts: true, ResourcePrices: true,
}

// CreateImport enregistre un import au statut draft avec ses lignes brutes.
func (s *OjoZone) CreateImport(ctx context.Context, createdBy string, input domain.ImportCreate) (domain.Import, error) {
	input.ResourceType = strings.TrimSpace(input.ResourceType)
	if !importableResources[input.ResourceType] {
		return domain.Import{}, ErrInvalidImport
	}
	if len(input.Rows) == 0 || len(input.Rows) > MaxImportRows {
		return domain.Import{}, ErrInvalidImport
	}
	for _, row := range input.Rows {
		if !json.Valid(row) {
			return domain.Import{}, ErrInvalidImport
		}
	}
	return s.repository.CreateImport(ctx, input, createdBy)
}

// ListImports retourne les imports paginés, du plus récent au plus ancien.
func (s *OjoZone) ListImports(ctx context.Context, pagination domain.Pagination) ([]domain.Import, error) {
	if err := validatePagination(pagination.Limit, pagination.Offset); err != nil {
		return nil, err
	}
	return s.repository.ListImports(ctx, pagination)
}

// GetImport retourne un import avec son état et son rapport.
func (s *OjoZone) GetImport(ctx context.Context, id string) (domain.Import, error) {
	if !validUUID(id) {
		return domain.Import{}, ErrInvalidID
	}
	return s.repository.GetImport(ctx, id)
}

// ValidateImport valide chaque ligne brutes et mémorise le rapport.
func (s *OjoZone) ValidateImport(ctx context.Context, id string) (domain.Import, error) {
	if !validUUID(id) {
		return domain.Import{}, ErrInvalidID
	}
	current, err := s.repository.GetImport(ctx, id)
	if err != nil {
		return domain.Import{}, err
	}
	if current.Status != domain.ImportStatusDraft {
		return domain.Import{}, ErrImportState
	}
	rows, err := s.repository.ListImportRows(ctx, id)
	if err != nil {
		return domain.Import{}, err
	}
	var report []domain.ImportReportEntry
	var validCount, invalidCount int32
	for _, row := range rows {
		raw := json.RawMessage(row.Payload)
		validationErr := validateImportResource(current.ResourceType, raw)
		if validationErr != nil {
			invalidCount++
			report = append(report, domain.ImportReportEntry{Line: row.LineNumber, Error: validationErr.Error()})
			if err := s.repository.SetImportRowStatus(ctx, id, row.LineNumber, false, &report[len(report)-1].Error); err != nil {
				return domain.Import{}, err
			}
			continue
		}
		validCount++
		if err := s.repository.SetImportRowStatus(ctx, id, row.LineNumber, true, nil); err != nil {
			return domain.Import{}, err
		}
	}
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return domain.Import{}, err
	}
	if err := s.repository.MarkImportValidated(ctx, id, validCount, invalidCount, reportJSON); err != nil {
		return domain.Import{}, err
	}
	return s.repository.GetImport(ctx, id)
}

// PublishImport applique les lignes valides via les upserts métier.
func (s *OjoZone) PublishImport(ctx context.Context, id string) (domain.Import, error) {
	if !validUUID(id) {
		return domain.Import{}, ErrInvalidID
	}
	current, err := s.repository.GetImport(ctx, id)
	if err != nil {
		return domain.Import{}, err
	}
	if current.Status != domain.ImportStatusValidated {
		return domain.Import{}, ErrImportState
	}
	rows, err := s.repository.ListImportRows(ctx, id)
	if err != nil {
		return domain.Import{}, err
	}
	var report []domain.ImportReportEntry
	for _, row := range rows {
		if !row.Valid {
			continue
		}
		raw := json.RawMessage(row.Payload)
		if publishErr := s.publishImportResource(ctx, current.ResourceType, raw); publishErr != nil {
			report = append(report, domain.ImportReportEntry{Line: row.LineNumber, Error: publishErr.Error()})
			if err := s.repository.SetImportRowStatus(ctx, id, row.LineNumber, false, &report[len(report)-1].Error); err != nil {
				return domain.Import{}, err
			}
		}
	}
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return domain.Import{}, err
	}
	if err := s.repository.MarkImportPublished(ctx, id, reportJSON); err != nil {
		return domain.Import{}, err
	}
	return s.repository.GetImport(ctx, id)
}

func validateImportResource(resource string, raw json.RawMessage) error {
	switch resource {
	case ResourceProducts:
		var value domain.ProductUpsert
		if json.Unmarshal(raw, &value) != nil {
			return ErrInvalidProduct
		}
		if err := normalizeAndValidateProduct(&value); err != nil {
			return err
		}
		if value.Barcode == nil && !validUUID(value.ID) {
			return ErrInvalidProduct
		}
		return nil
	case ResourcePrices:
		var value domain.ProductPriceUpsert
		if json.Unmarshal(raw, &value) != nil {
			return ErrInvalidPrice
		}
		if err := normalizeAndValidateProductPrice(&value); err != nil {
			return err
		}
		if value.SourceRecordID == nil {
			return ErrInvalidPrice
		}
		return nil
	default:
		_, err := decodeAndValidateResource(resource, raw, false, domain.ResourceKey{})
		return err
	}
}

func (s *OjoZone) publishImportResource(ctx context.Context, resource string, raw json.RawMessage) error {
	switch resource {
	case ResourceProducts:
		var value domain.ProductUpsert
		if json.Unmarshal(raw, &value) != nil {
			return ErrInvalidProduct
		}
		_, err := s.UpsertProduct(ctx, value)
		return err
	case ResourcePrices:
		var value domain.ProductPriceUpsert
		if json.Unmarshal(raw, &value) != nil {
			return ErrInvalidPrice
		}
		_, err := s.UpsertProductPrice(ctx, value)
		return err
	default:
		_, err := s.UpsertResource(ctx, resource, raw)
		return err
	}
}
