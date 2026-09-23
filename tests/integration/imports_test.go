package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/laurentpoirierfr/ojozone/tests/internal/api"
)

func createImportAsAdmin(t *testing.T, admin api.AuthResult, resourceType string, rows ...map[string]any) api.Import {
	t.Helper()
	rawRows := make([]json.RawMessage, 0, len(rows))
	for _, row := range rows {
		encoded, err := json.Marshal(row)
		if err != nil {
			t.Fatalf("serialisation ligne : %v", err)
		}
		rawRows = append(rawRows, encoded)
	}
	created, problem, err := client.CreateImport(context.Background(), admin.Tokens.AccessToken, resourceType, rawRows)
	if err != nil {
		t.Fatalf("transport creation import : %v", err)
	}
	requireNoProblem(t, problem, "creation import")
	if created.ID == "" || created.Status != "draft" || int(created.LineCount) != len(rows) {
		t.Fatalf("import cree inattendu : %+v", created)
	}
	return created
}

func TestAdminCreatesAndValidatesImport(t *testing.T) {
	admin := adminSession(t)
	suffix := fmt.Sprintf("g%d", time.Now().UnixNano()%100000000)
	imported := createImportAsAdmin(t, admin, "units",
		map[string]any{"code": suffix, "dimension": "mass", "to_base_factor": "0.001"},
		map[string]any{"code": "bad-unit", "dimension": "inexistante", "to_base_factor": "1"},
	)

	validated, problem, err := client.ValidateImport(context.Background(), admin.Tokens.AccessToken, imported.ID)
	if err != nil {
		t.Fatalf("transport validation import : %v", err)
	}
	requireNoProblem(t, problem, "validation import")
	if validated.Status != "validated" || int(validated.ValidCount) != 1 || int(validated.InvalidCount) != 1 {
		t.Fatalf("rapport de validation inattendu : %+v", validated)
	}
	if len(validated.Report) != 1 || validated.Report[0].Line != 2 {
		t.Fatalf("rapport d'erreurs inattendu : %+v", validated.Report)
	}
}

func TestAdminPublishesValidImportAndDataAppears(t *testing.T) {
	admin := adminSession(t)
	code := fmt.Sprintf("pub%d", time.Now().UnixNano()%100000000)
	imported := createImportAsAdmin(t, admin, "units",
		map[string]any{"code": code, "dimension": "volume", "to_base_factor": "0.001"},
	)

	validated, problem, err := client.ValidateImport(context.Background(), admin.Tokens.AccessToken, imported.ID)
	if err != nil {
		t.Fatalf("transport validation : %v", err)
	}
	requireNoProblem(t, problem, "validation")
	if validated.ValidCount != 1 {
		t.Fatalf("validation inattendue : %+v", validated)
	}

	published, problem, err := client.PublishImport(context.Background(), admin.Tokens.AccessToken, imported.ID)
	if err != nil {
		t.Fatalf("transport publication : %v", err)
	}
	requireNoProblem(t, problem, "publication")
	if published.Status != "published" {
		t.Fatalf("statut apres publication : %+v", published)
	}

	units, _, problem, err := client.ListUnits(context.Background())
	if err != nil {
		t.Fatalf("transport liste des unites : %v", err)
	}
	requireNoProblem(t, problem, "liste des unites")
	found := false
	for _, unit := range units {
		if unit.Code == code {
			found = true
		}
	}
	if !found {
		t.Fatalf("unite %q absente apres publication", code)
	}
}

func TestAdminPublishRejectsDraftImport(t *testing.T) {
	admin := adminSession(t)
	imported := createImportAsAdmin(t, admin, "units",
		map[string]any{"code": fmt.Sprintf("d%d", time.Now().UnixNano()%100000000), "dimension": "mass", "to_base_factor": 1},
	)
	_, problem, err := client.PublishImport(context.Background(), admin.Tokens.AccessToken, imported.ID)
	if err != nil {
		t.Fatalf("transport publication prematuree : %v", err)
	}
	requireProblem(t, problem, 409, "import_state", "publication prematuree")
}

func TestAdminCannotCreateImportWithUnknownResource(t *testing.T) {
	admin := adminSession(t)
	row, _ := json.Marshal(map[string]any{"code": "x"})
	_, problem, err := client.CreateImport(context.Background(), admin.Tokens.AccessToken, "widgets", []json.RawMessage{row})
	if err != nil {
		t.Fatalf("transport import inconnu : %v", err)
	}
	requireProblem(t, problem, 400, "invalid_request", "type de ressource inconnu")
}

func TestMemberCannotManageImportsThroughAPI(t *testing.T) {
	member := registerMember(t)
	row, _ := json.Marshal(map[string]any{"code": "x", "dimension": "mass", "to_base_factor": 1})
	_, problem, err := client.CreateImport(context.Background(), member.Tokens.AccessToken, "units", []json.RawMessage{row})
	if err != nil {
		t.Fatalf("transport import membre : %v", err)
	}
	requireProblem(t, problem, 403, "forbidden", "import membre")
}