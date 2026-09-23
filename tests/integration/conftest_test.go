// Package integration verifie l'API OjoZone de bout en bout contre une base
// PostgreSQL réelle. La suite suppose qu'une instance de l'API tourne
// (scripts/run.sh) et décrite dans tests/README.md.
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/laurentpoirierfr/ojozone/tests/internal/api"
)

// decodeJSON lit le corps brut d'une reponse non consumeé.
func decodeJSON(response *http.Response, out any) error {
	defer response.Body.Close()
	return json.NewDecoder(response.Body).Decode(out)
}

var client *api.Client

// TestMain verifie que le serveur est joignable avant d'executer la suite.
func TestMain(m *testing.M) {
	baseURL := getenv("OJZONE_TEST_BASE_URL", "http://127.0.0.1:18081")
	client = api.New(baseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/ops/liveness", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "construction de la requête de vivacité : %v\n", err)
		os.Exit(1)
	}
	response, err := client.HTTP.Do(request)
	cancel()
	if err != nil || response != nil && response.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "API inaccessible sur %s. Lancez d'abord : make -C tests check\n", baseURL)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func getenv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

// uniqueEmail construit une adresse electronique unique pour un test donne.
func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@example.com", prefix, time.Now().UnixNano())
}

// registerMember cree un compte membre frais et le retourne avec ses jetons.
func registerMember(t *testing.T) api.AuthResult {
	t.Helper()
	result, problem, err := client.Register(context.Background(), uniqueEmail("e2e"), "mot-de-passe-123", "fr")
	if err != nil {
		t.Fatalf("transport pendant l'inscription : %v", err)
	}
	requireNoProblem(t, problem, "inscription")
	if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" {
		t.Fatal("inscription sans paire de jetons")
	}
	return result
}

// adminSession connecte le compte administrateur prepare par scripts/run.sh.
func adminSession(t *testing.T) api.AuthResult {
	t.Helper()
	email := getenv("OJZONE_TEST_ADMIN_EMAIL", "admin@example.com")
	password := getenv("OJZONE_TEST_ADMIN_PASSWORD", "admin-password-123")
	result, problem, err := client.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("transport pendant la connexion admin : %v", err)
	}
	requireNoProblem(t, problem, "connexion admin")
	return result
}

// moderatorSession connecte le compte moderateur prepare par scripts/run.sh.
func moderatorSession(t *testing.T) api.AuthResult {
	t.Helper()
	email := getenv("OJZONE_TEST_MODERATOR_EMAIL", "moderator@example.com")
	password := getenv("OJZONE_TEST_MODERATOR_PASSWORD", "moderator-password-123")
	result, problem, err := client.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("transport pendant la connexion moderateur : %v", err)
	}
	requireNoProblem(t, problem, "connexion moderateur")
	return result
}

// requireNoProblem echoue si un reponse d'erreur HTTP a ete renvoyee.
func requireNoProblem(t *testing.T, problem *api.Problem, step string) {
	t.Helper()
	if problem != nil {
		t.Fatalf("%s : erreur API %d %s (%s)", step, problem.Status, problem.Type, problem.Detail)
	}
}

// requireProblem verifie le type et le statut d'une reponse d'erreur.
func requireProblem(t *testing.T, problem *api.Problem, status int, problemType string, step string) {
	t.Helper()
	if problem == nil {
		t.Fatalf("%s : aucune erreur remontee, statut attendu %d", step, status)
	}
	if problem.Status != status {
		t.Fatalf("%s : statut obtenu %d, attendu %d", step, problem.Status, status)
	}
	if problemType != "" && problem.Type != problemType {
		t.Fatalf("%s : type obtenu %q, attendu %q", step, problem.Type, problemType)
	}
	if problem.Status != problem.StatusCode() {
		t.Fatalf("%s : Status et StatusCode incohérents", step)
	}
}
