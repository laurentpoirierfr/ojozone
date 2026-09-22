package integration

import (
	"context"
	"net/http"
	"testing"
)

const (
	fuelTypeID = "b25fbc0e-9d7a-4d2e-bf5f-7d2a1f0a3c92"
	observedAt = "2026-09-22T12:00:00Z"
)

func productContributionPayload(productID string) map[string]any {
	return map[string]any{
		"product_id":   productID,
		"location_id":  locationID,
		"amount":       "2.35",
		"currency":     "EUR",
		"quantity":     "1.000",
		"unit_code":    "L",
		"is_promotion": false,
		"observed_at":  observedAt,
	}
}

func fuelContributionPayload() map[string]any {
	return map[string]any{
		"fuel_type_id":     fuelTypeID,
		"location_id":      locationID,
		"amount_per_litre": "1.849",
		"currency":         "EUR",
		"observed_at":      observedAt,
	}
}

func TestMemberSubmitsProductContribution(t *testing.T) {
	admin := adminSession(t)
	productID := createProductAsAdmin(t, admin, uniqueBarcode())
	member := registerMember(t)

	result, problem, err := client.SubmitProductContribution(context.Background(), member.Tokens.AccessToken, productContributionPayload(productID))
	if err != nil {
		t.Fatalf("transport soumission : %v", err)
	}
	requireNoProblem(t, problem, "soumission produit")
	if result.ID == "" || result.Status != "pending" {
		t.Fatalf("resultat inattendu : %+v", result)
	}
	if len(result.Checks) == 0 {
		t.Fatal("aucun controle renvoye")
	}

	detail, problem, err := client.GetContribution(context.Background(), member.Tokens.AccessToken, result.ID)
	if err != nil {
		t.Fatalf("transport lecture : %v", err)
	}
	requireNoProblem(t, problem, "lecture contribution")
	if detail.Type != "product_price" || detail.Status != "pending" {
		t.Fatalf("details inattendues : %+v", detail)
	}
	if detail.ContributorID != member.User.ID {
		t.Fatalf("auteur %q, attendu %q", detail.ContributorID, member.User.ID)
	}
	if detail.Source.Kind != "community" {
		t.Fatalf("source %q, attendue community", detail.Source.Kind)
	}
	if detail.Amount != "2.35" || detail.Quantity == nil || *detail.Quantity != "1.000" {
		t.Fatalf("montant ou quantite inattendus : %+v", detail)
	}

	contributions, problem, err := client.MyContributions(context.Background(), member.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("transport suivi : %v", err)
	}
	requireNoProblem(t, problem, "suivi des contributions")
	found := false
	for _, contribution := range contributions {
		if contribution.ID == result.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("contribution absente du suivi /me/contributions")
	}
}

func TestMemberSubmitsFuelContribution(t *testing.T) {
	member := registerMember(t)
	result, problem, err := client.SubmitFuelContribution(context.Background(), member.Tokens.AccessToken, fuelContributionPayload())
	if err != nil {
		t.Fatalf("transport soumission : %v", err)
	}
	requireNoProblem(t, problem, "soumission carburant")
	if result.Status != "pending" {
		t.Fatalf("statut inattendu : %s", result.Status)
	}

	detail, problem, err := client.GetContribution(context.Background(), member.Tokens.AccessToken, result.ID)
	if err != nil {
		t.Fatalf("transport lecture : %v", err)
	}
	requireNoProblem(t, problem, "lecture carburant")
	if detail.Type != "fuel_price" || detail.Amount != "1.8490" {
		t.Fatalf("details inattendues : %+v", detail)
	}
	if detail.Subject != "SP95-E10" {
		t.Fatalf("sujet %q, attendu SP95-E10", detail.Subject)
	}
}

func TestContributionsRequireAuthentication(t *testing.T) {
	response, err := client.Raw(context.Background(), http.MethodPost, "/api/v1/contributions/product-prices", "", []byte(`{}`))
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("statut obtenu %d, attendu %d", response.StatusCode, http.StatusUnauthorized)
	}

	response, err = client.Raw(context.Background(), http.MethodGet, "/api/v1/contributions/2c65c9b4-4527-4e2a-9d2a-6b0a7a1f8c11", "", nil)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("statut obtenu %d, attendu %d", response.StatusCode, http.StatusUnauthorized)
	}
}

func TestContributionValidationIsRejected(t *testing.T) {
	member := registerMember(t)
	payload := productContributionPayload("2c65c9b4-4527-4e2a-9d2a-6b0a7a1f8c11")
	delete(payload, "unit_code")
	_, problem, err := client.SubmitProductContribution(context.Background(), member.Tokens.AccessToken, payload)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusBadRequest, "invalid_contribution", "unite manquante")
}

func TestContributionIsPrivateToOwner(t *testing.T) {
	admin := adminSession(t)
	productID := createProductAsAdmin(t, admin, uniqueBarcode())
	owner := registerMember(t)
	other := registerMember(t)

	result, problem, err := client.SubmitProductContribution(context.Background(), owner.Tokens.AccessToken, productContributionPayload(productID))
	if err != nil {
		t.Fatalf("transport soumission : %v", err)
	}
	requireNoProblem(t, problem, "soumission")

	_, problem, err = client.GetContribution(context.Background(), other.Tokens.AccessToken, result.ID)
	if err != nil {
		t.Fatalf("transport lecture : %v", err)
	}
	requireProblem(t, problem, http.StatusForbidden, "forbidden", "lecture par un autre membre")

	_, problem, err = client.PatchContribution(context.Background(), other.Tokens.AccessToken, result.ID, map[string]any{"amount": "2.45"})
	if err != nil {
		t.Fatalf("transport correction : %v", err)
	}
	requireProblem(t, problem, http.StatusForbidden, "forbidden", "correction par un autre membre")

	problem, err = client.DeleteContribution(context.Background(), other.Tokens.AccessToken, result.ID)
	if err != nil {
		t.Fatalf("transport retrait : %v", err)
	}
	requireProblem(t, problem, http.StatusForbidden, "forbidden", "retrait par un autre membre")
}

func TestOwnerCanPatchContribution(t *testing.T) {
	admin := adminSession(t)
	productID := createProductAsAdmin(t, admin, uniqueBarcode())
	member := registerMember(t)

	result, problem, err := client.SubmitProductContribution(context.Background(), member.Tokens.AccessToken, productContributionPayload(productID))
	if err != nil {
		t.Fatalf("transport soumission : %v", err)
	}
	requireNoProblem(t, problem, "soumission")

	updated, problem, err := client.PatchContribution(context.Background(), member.Tokens.AccessToken, result.ID, map[string]any{"amount": "2.75", "quantity": "1.000"})
	if err != nil {
		t.Fatalf("transport correction : %v", err)
	}
	requireNoProblem(t, problem, "correction")
	if updated.Amount != "2.75" {
		t.Fatalf("montant non corrige : %+v", updated)
	}

	detail, problem, err := client.GetContribution(context.Background(), member.Tokens.AccessToken, result.ID)
	if err != nil {
		t.Fatalf("transport relecture : %v", err)
	}
	requireNoProblem(t, problem, "relecture")
	if detail.Amount != "2.75" {
		t.Fatalf("montant persisté : %+v", detail)
	}
}

func TestOwnerCanWithdrawContribution(t *testing.T) {
	admin := adminSession(t)
	productID := createProductAsAdmin(t, admin, uniqueBarcode())
	member := registerMember(t)

	result, problem, err := client.SubmitProductContribution(context.Background(), member.Tokens.AccessToken, productContributionPayload(productID))
	if err != nil {
		t.Fatalf("transport soumission : %v", err)
	}
	requireNoProblem(t, problem, "soumission")

	problem, err = client.DeleteContribution(context.Background(), member.Tokens.AccessToken, result.ID)
	if err != nil {
		t.Fatalf("transport retrait : %v", err)
	}
	requireNoProblem(t, problem, "retrait")

	_, problem, err = client.GetContribution(context.Background(), member.Tokens.AccessToken, result.ID)
	if err != nil {
		t.Fatalf("transport lecture : %v", err)
	}
	requireProblem(t, problem, http.StatusNotFound, "not_found", "lecture apres retrait")
}

func TestUnknownContributionIsNotFound(t *testing.T) {
	member := registerMember(t)
	_, problem, err := client.GetContribution(context.Background(), member.Tokens.AccessToken, "2c65c9b4-4527-4e2a-9d2a-6b0a7a1f8c11")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusNotFound, "not_found", "contribution inconnue")
}
