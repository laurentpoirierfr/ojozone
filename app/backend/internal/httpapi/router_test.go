package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

type serviceStub struct {
	readinessError error
}

func (s serviceStub) Readiness(context.Context) error { return s.readinessError }
func (s serviceStub) ListProducts(context.Context, domain.ProductFilter) ([]domain.Product, error) {
	return []domain.Product{}, nil
}
func (s serviceStub) GetProduct(context.Context, string) (domain.Product, error) {
	return domain.Product{}, domain.ErrNotFound
}
func (s serviceStub) GetProductByBarcode(context.Context, string) (domain.Product, error) {
	return domain.Product{}, domain.ErrNotFound
}
func (s serviceStub) UpsertProduct(context.Context, domain.ProductUpsert) (domain.Product, error) {
	return domain.Product{}, nil
}
func (s serviceStub) ReplaceProduct(context.Context, string, domain.ProductUpsert) (domain.Product, error) {
	return domain.Product{}, nil
}
func (s serviceStub) DeleteProduct(context.Context, string) error { return nil }
func (s serviceStub) ListProductPrices(context.Context, string, domain.PriceFilter) ([]domain.ProductPrice, error) {
	return []domain.ProductPrice{}, nil
}
func (s serviceStub) ListAllProductPrices(context.Context, domain.PriceFilter) ([]domain.ProductPrice, error) {
	return []domain.ProductPrice{}, nil
}
func (s serviceStub) GetProductPrice(context.Context, string) (domain.ProductPrice, error) {
	return domain.ProductPrice{}, nil
}
func (s serviceStub) UpsertProductPrice(context.Context, domain.ProductPriceUpsert) (domain.ProductPrice, error) {
	return domain.ProductPrice{}, nil
}
func (s serviceStub) ReplaceProductPrice(context.Context, string, domain.ProductPriceUpsert) (domain.ProductPrice, error) {
	return domain.ProductPrice{}, nil
}
func (s serviceStub) DeleteProductPrice(context.Context, string) error { return nil }
func (s serviceStub) ListResource(context.Context, string, domain.Pagination) (any, error) {
	return []any{}, nil
}
func (s serviceStub) GetResource(context.Context, string, domain.ResourceKey) (any, error) {
	return struct{}{}, nil
}
func (s serviceStub) UpsertResource(context.Context, string, json.RawMessage) (any, error) {
	return struct{}{}, nil
}
func (s serviceStub) ReplaceResource(context.Context, string, domain.ResourceKey, json.RawMessage) (any, error) {
	return struct{}{}, nil
}
func (s serviceStub) DeleteResource(context.Context, string, domain.ResourceKey) error { return nil }

func TestOperationalRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name     string
		path     string
		service  serviceStub
		expected int
	}{
		{name: "liveness", path: "/ops/liveness", expected: http.StatusOK},
		{name: "readiness", path: "/ops/readiness", expected: http.StatusOK},
		{name: "not ready", path: "/ops/readiness", service: serviceStub{readinessError: errors.New("database unavailable")}, expected: http.StatusServiceUnavailable},
		{name: "infos", path: "/ops/infos", expected: http.StatusOK},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			NewRouter(test.service, nil, BuildInfo{Name: "ojozone-api"}).ServeHTTP(response, request)
			if response.Code != test.expected {
				t.Fatalf("status obtenu %d, attendu %d", response.Code, test.expected)
			}
		})
	}
}

func TestAPIRoutesRemainUnderV1(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(serviceStub{}, nil, BuildInfo{})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/products", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status obtenu %d, attendu %d", response.Code, http.StatusOK)
	}

	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/products", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("route non versionnée : status obtenu %d, attendu %d", response.Code, http.StatusNotFound)
	}
}

func TestStaticIndexIsServedWithoutRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	staticFS := fstest.MapFS{"index.html": {Data: []byte("<h1>OjoZone</h1>")}}
	response := httptest.NewRecorder()
	NewRouter(serviceStub{}, staticFS, BuildInfo{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status obtenu %d, attendu %d", response.Code, http.StatusOK)
	}
}

func TestProductWriteRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(serviceStub{}, nil, BuildInfo{})
	body := []byte(`{"category_id":"765c4c3e-9a2e-4a7f-a272-586a311cbb80","name":"Lait","barcode":"3017620422003","reference_unit":"L","attributes":{}}`)
	tests := []struct {
		method   string
		path     string
		body     []byte
		expected int
	}{
		{method: http.MethodPost, path: "/api/v1/products", body: body, expected: http.StatusOK},
		{method: http.MethodPut, path: "/api/v1/products/ec2d9232-7ec5-44c4-85fc-1dfdd80b9d31", body: body, expected: http.StatusOK},
		{method: http.MethodDelete, path: "/api/v1/products/ec2d9232-7ec5-44c4-85fc-1dfdd80b9d31", expected: http.StatusNoContent},
		{method: http.MethodPost, path: "/api/v1/products", body: []byte(`{"name":`), expected: http.StatusBadRequest},
	}

	for _, test := range tests {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(test.method, test.path, bytes.NewReader(test.body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)
		if response.Code != test.expected {
			t.Errorf("%s %s : status obtenu %d, attendu %d", test.method, test.path, response.Code, test.expected)
		}
	}
}

func TestProductPriceRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(serviceStub{}, nil, BuildInfo{})
	priceID := "81dcce22-c2d6-4316-b4ba-61f256436e86"
	body := []byte(`{"product_id":"ec2d9232-7ec5-44c4-85fc-1dfdd80b9d31","location_id":"9a80c20f-9055-442d-97d1-43ee336230f0","source_id":"0cbdc6bf-361b-4878-b407-e77f735098af","amount":"2.35","currency":"EUR","quantity":"1.000","unit_code":"L","normalized_amount":"2.3500","observed_at":"2026-09-22T12:00:00Z","status":"approved","source_record_id":"provider-row-1"}`)
	tests := []struct {
		method   string
		path     string
		body     []byte
		expected int
	}{
		{method: http.MethodPost, path: "/api/v1/product-prices", body: body, expected: http.StatusOK},
		{method: http.MethodGet, path: "/api/v1/product-prices/" + priceID, expected: http.StatusOK},
		{method: http.MethodPut, path: "/api/v1/product-prices/" + priceID, body: body, expected: http.StatusOK},
		{method: http.MethodDelete, path: "/api/v1/product-prices/" + priceID, expected: http.StatusNoContent},
	}

	for _, test := range tests {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(test.method, test.path, bytes.NewReader(test.body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)
		if response.Code != test.expected {
			t.Errorf("%s %s : status obtenu %d, attendu %d", test.method, test.path, response.Code, test.expected)
		}
	}
}

func TestRemainingAPIRoutesAreRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routes := NewRouter(serviceStub{}, nil, BuildInfo{}).Routes()
	present := make(map[string]bool, len(routes))
	for _, route := range routes {
		present[route.Method+" "+route.Path] = true
	}
	crud := []string{"geo-areas", "sources", "categories", "units", "merchants", "locations", "fuel-types", "fuel-prices", "housing-observations", "income-observations", "evidence-files"}
	for _, resource := range crud {
		key := ":id"
		if resource == "units" {
			key = ":code"
		}
		for _, entry := range []string{http.MethodGet + " /api/v1/" + resource, http.MethodPost + " /api/v1/" + resource, http.MethodGet + " /api/v1/" + resource + "/" + key, http.MethodPut + " /api/v1/" + resource + "/" + key, http.MethodDelete + " /api/v1/" + resource + "/" + key} {
			if !present[entry] {
				t.Errorf("route absente : %s", entry)
			}
		}
	}
	aggregatePath := "/api/v1/price-aggregates/:geo_area_id/:metric_type/:subject_id/:month"
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		if !present[method+" "+aggregatePath] {
			t.Errorf("route agrégat absente : %s", method)
		}
	}
	for _, entry := range []string{"GET /api/v1/moderation-events", "POST /api/v1/moderation-events", "GET /api/v1/moderation-events/:id", "GET /api/v1/admin/users", "GET /api/v1/admin/users/:id"} {
		if !present[entry] {
			t.Errorf("route absente : %s", entry)
		}
	}
	for _, entry := range []string{"PUT /api/v1/moderation-events/:id", "DELETE /api/v1/moderation-events/:id", "POST /api/v1/admin/users", "PUT /api/v1/admin/users/:id", "DELETE /api/v1/admin/users/:id"} {
		if present[entry] {
			t.Errorf("route interdite enregistrée : %s", entry)
		}
	}
}
