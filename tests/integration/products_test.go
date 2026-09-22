package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/laurentpoirierfr/ojozone/tests/internal/api"
)

// seedCategory et les autres UUID font reference aux fixtures installees par
// tests/scripts/run.sh dans la base ojozone_test.
const (
	categoryID = "765c4c3e-9a2e-4a7f-a272-586a311cbb80"
	sourceID   = "0cbdc6bf-361b-4878-b407-e77f735098af"
	locationID = "9a80c20f-9055-442d-97d1-43ee336230f0"
)

func validProductPayload(barcode string) map[string]any {
	return map[string]any{
		"category_id":    categoryID,
		"name":           "Lait entier",
		"barcode":        barcode,
		"reference_unit": "L",
		"is_generic":     false,
		"attributes":     map[string]any{"brand": "demo"},
	}
}

func validPricePayload(productID, sourceRecordID string) map[string]any {
	return map[string]any{
		"product_id":        productID,
		"location_id":       locationID,
		"source_id":         sourceID,
		"amount":            "2.35",
		"currency":          "EUR",
		"quantity":          "1.000",
		"unit_code":         "L",
		"normalized_amount": "2.3500",
		"is_promotion":      false,
		"observed_at":       "2026-09-22T12:00:00Z",
		"source_record_id":  sourceRecordID,
		"status":            "approved",
	}
}

func uniqueBarcode() string {
	return fmt.Sprintf("301762%d", time.Now().UnixNano()%100000000)
}

func createProductAsAdmin(t *testing.T, admin api.AuthResult, barcode string) string {
	t.Helper()
	product, problem, err := client.CreateProduct(context.Background(), admin.Tokens.AccessToken, validProductPayload(barcode))
	if err != nil {
		t.Fatalf("transport creation produit : %v", err)
	}
	requireNoProblem(t, problem, "creation produit")
	if product.ID == "" {
		t.Fatal("produit cree sans identifiant")
	}
	return product.ID
}

func TestAdminCreatesAndFetchesProduct(t *testing.T) {
	admin := adminSession(t)
	productID := createProductAsAdmin(t, admin, uniqueBarcode())

	product, problem, err := client.GetProduct(context.Background(), productID)
	if err != nil {
		t.Fatalf("transport lecture produit : %v", err)
	}
	requireNoProblem(t, problem, "lecture produit")
	if product.Name != "Lait entier" || product.ReferenceUnit != "L" {
		t.Fatalf("produit inattendu : %+v", product)
	}
}

func TestProductPriceUpsertIsIdempotent(t *testing.T) {
	admin := adminSession(t)
	productID := createProductAsAdmin(t, admin, uniqueBarcode())
	recordID := fmt.Sprintf("e2e-%d", time.Now().UnixNano())

	first, problem, err := client.CreateProductPrice(context.Background(), admin.Tokens.AccessToken, validPricePayload(productID, recordID))
	if err != nil {
		t.Fatalf("transport prix : %v", err)
	}
	requireNoProblem(t, problem, "creation prix")
	if first.ID == "" || first.Amount != "2.35" {
		t.Fatalf("prix inattendu : %+v", first)
	}

	second, problem, err := client.CreateProductPrice(context.Background(), admin.Tokens.AccessToken, validPricePayload(productID, recordID))
	if err != nil {
		t.Fatalf("transport prix (2e) : %v", err)
	}
	requireNoProblem(t, problem, "rejeu prix")
	if second.ID != first.ID {
		t.Fatalf("l'upsert par source_record_id doit etre idempotent : %s != %s", first.ID, second.ID)
	}
}

func TestProductPriceAppearsInProductPrices(t *testing.T) {
	admin := adminSession(t)
	productID := createProductAsAdmin(t, admin, uniqueBarcode())
	price, problem, err := client.CreateProductPrice(context.Background(), admin.Tokens.AccessToken, validPricePayload(productID, fmt.Sprintf("e2e-%d", time.Now().UnixNano())))
	if err != nil {
		t.Fatalf("transport prix : %v", err)
	}
	requireNoProblem(t, problem, "creation prix")

	prices, meta, problem, err := client.ListProductPrices(context.Background(), productID)
	if err != nil {
		t.Fatalf("transport liste prix : %v", err)
	}
	requireNoProblem(t, problem, "liste prix")
	if meta.Limit <= 0 {
		t.Fatalf("meta de pagination inattendu : %+v", meta)
	}
	found := false
	for _, candidate := range prices {
		if candidate.ID == price.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("prix %s absent de la liste %+v", price.ID, prices)
	}
}

func TestMemberCannotWriteCatalog(t *testing.T) {
	member := registerMember(t)
	_, problem, err := client.CreateProduct(context.Background(), member.Tokens.AccessToken, validProductPayload(uniqueBarcode()))
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusForbidden, "forbidden", "ecriture membre")
}

func TestAnonymousCannotWriteCatalog(t *testing.T) {
	_, problem, err := client.CreateProduct(context.Background(), "", validProductPayload(uniqueBarcode()))
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusUnauthorized, "unauthorized", "ecriture anonyme")
}

func TestPublicProductSearch(t *testing.T) {
	code := uniqueBarcode()
	admin := adminSession(t)
	name := fmt.Sprintf("Yaourt nature %d", time.Now().UnixNano()%100000)
	payload := validProductPayload(code)
	payload["name"] = name
	if _, problem, err := client.CreateProduct(context.Background(), admin.Tokens.AccessToken, payload); err != nil {
		t.Fatalf("transport : %v", err)
	} else {
		requireNoProblem(t, problem, "creation produit")
	}

	products, meta, problem, err := client.ListProducts(context.Background(), "yaourt")
	if err != nil {
		t.Fatalf("transport recherche : %v", err)
	}
	requireNoProblem(t, problem, "recherche publique")
	if meta.Limit <= 0 {
		t.Fatalf("meta inattendu : %+v", meta)
	}
	found := false
	for _, product := range products {
		if product.Name == name {
			found = true
		}
	}
	if !found {
		t.Fatalf("produit %q introuvable dans %+v", name, products)
	}
}
