package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/laurentpoirierfr/ojozone/tests/internal/api"
)

func TestUnknownRouteReturnsProblemJSON(t *testing.T) {
	response, err := client.Raw(context.Background(), http.MethodGet, "/api/v1/inexistante", "", nil)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("statut obtenu %d, attendu %d", response.StatusCode, http.StatusNotFound)
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "application/problem+json" {
		t.Fatalf("Content-Type inattendu : %q", contentType)
	}
	var problem api.Problem
	if err := decodeJSON(response, &problem); err != nil {
		t.Fatalf("decodage problem : %v", err)
	}
	if problem.Type != "not_found" || problem.Status != http.StatusNotFound {
		t.Fatalf("problem inattendu : %+v", problem)
	}
}

func TestMalformedJSONBodyIsRejected(t *testing.T) {
	response, err := client.Raw(context.Background(), http.MethodPost, "/api/v1/auth/login", "", []byte(`{"email":`))
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	var problem api.Problem
	if err := decodeJSON(response, &problem); err != nil {
		t.Fatalf("decodage problem : %v", err)
	}
	if problem.Type != "invalid_json" || problem.Status != http.StatusBadRequest {
		t.Fatalf("problem inattendu : %+v", problem)
	}
}

func TestUnknownProductReturns404(t *testing.T) {
	_, problem, err := client.GetProduct(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusNotFound, "not_found", "produit inconnu")
}

func TestAdminUsersRouteIsRoleProtected(t *testing.T) {
	_, problem, err := client.ListAdminUsers(context.Background(), "")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusUnauthorized, "unauthorized", "admin sans jeton")

	member := registerMember(t)
	_, problem, err = client.ListAdminUsers(context.Background(), member.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusForbidden, "forbidden", "admin en membre")

	admin := adminSession(t)
	users, problem, err := client.ListAdminUsers(context.Background(), admin.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireNoProblem(t, problem, "admin en admin")
	if len(users) == 0 {
		t.Fatal("la liste des utilisateurs ne doit pas etre vide (admin saisi)")
	}
	for _, user := range users {
		if user.ID == "" || user.Email == "" {
			t.Fatalf("utilisateur incomplet exposé : %+v", user)
		}
	}
}

func TestModerationEventsRequiresModeratorOrAdmin(t *testing.T) {
	member := registerMember(t)
	payload := []byte(`{"observation_type":"product_price","observation_id":"00000000-0000-0000-0000-000000000000","new_status":"rejected","reason_code":"test","note":"validation e2e"}`)
	response, err := client.Raw(context.Background(), http.MethodPost, "/api/v1/moderation-events", member.Tokens.AccessToken, payload)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	var problem api.Problem
	if err := decodeJSON(response, &problem); err != nil {
		t.Fatalf("decodage problem : %v", err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("statut obtenu %d, attendu %d (member interdit de moderer)", response.StatusCode, http.StatusForbidden)
	}
	if problem.Type != "forbidden" {
		t.Fatalf("problem inattendu : %+v", problem)
	}
}
