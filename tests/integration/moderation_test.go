package integration

import (
	"context"
	"net/http"
	"testing"
)

func submitProductContribution(t *testing.T, accessToken, productID string) string {
	t.Helper()
	result, problem, err := client.SubmitProductContribution(context.Background(), accessToken, productContributionPayload(productID))
	if err != nil {
		t.Fatalf("transport soumission produit : %v", err)
	}
	requireNoProblem(t, problem, "soumission produit")
	if result.Status != "pending" || result.ID == "" {
		t.Fatalf("soumission inattendue : %+v", result)
	}
	return result.ID
}

func submitFuelContribution(t *testing.T, accessToken string) string {
	t.Helper()
	result, problem, err := client.SubmitFuelContribution(context.Background(), accessToken, fuelContributionPayload())
	if err != nil {
		t.Fatalf("transport soumission carburant : %v", err)
	}
	requireNoProblem(t, problem, "soumission carburant")
	return result.ID
}

func TestModeratorSeesPendingQueue(t *testing.T) {
	admin := adminSession(t)
	productID := createProductAsAdmin(t, admin, uniqueBarcode())
	member := registerMember(t)
	productContribution := submitProductContribution(t, member.Tokens.AccessToken, productID)
	fuelContribution := submitFuelContribution(t, member.Tokens.AccessToken)

	moderator := moderatorSession(t)
	items, meta, problem, err := client.ListModerationQueue(context.Background(), moderator.Tokens.AccessToken, "pending")
	if err != nil {
		t.Fatalf("transport file : %v", err)
	}
	requireNoProblem(t, problem, "file de moderation")
	if meta.Limit <= 0 {
		t.Fatalf("meta inattendu : %+v", meta)
	}
	seenProduct, seenFuel := false, false
	for _, item := range items {
		if item.ID == productContribution && item.Type == "product_price" && item.Subject == "Lait entier" {
			seenProduct = true
		}
		if item.ID == fuelContribution && item.Type == "fuel_price" && item.Amount == "1.8490" {
			seenFuel = true
		}
	}
	if !seenProduct || !seenFuel {
		t.Fatalf("file incomplete (produit %v, carburant %v) : %+v", seenProduct, seenFuel, items)
	}
}

func TestModeratorApprovesProductContribution(t *testing.T) {
	admin := adminSession(t)
	productID := createProductAsAdmin(t, admin, uniqueBarcode())
	member := registerMember(t)
	contributionID := submitProductContribution(t, member.Tokens.AccessToken, productID)

	moderator := moderatorSession(t)
	detail, problem, err := client.ApproveContribution(context.Background(), moderator.Tokens.AccessToken, contributionID, "Prix cohérent")
	if err != nil {
		t.Fatalf("transport approbation : %v", err)
	}
	requireNoProblem(t, problem, "approbation")
	if detail.Status != "approved" {
		t.Fatalf("statut obtenu %s, attendu approved", detail.Status)
	}

	contributorDetail, problem, err := client.GetContribution(context.Background(), member.Tokens.AccessToken, contributionID)
	if err != nil {
		t.Fatalf("transport lecteur contributeur : %v", err)
	}
	requireNoProblem(t, problem, "lecture contributeur")
	if contributorDetail.Status != "approved" {
		t.Fatalf("statut contributeur %s, attendu approved", contributorDetail.Status)
	}

	events, _, problem, err := client.ListModerationEvents(context.Background())
	if err != nil {
		t.Fatalf("transport audit : %v", err)
	}
	requireNoProblem(t, problem, "audit de moderation")
	found := false
	for _, event := range events {
		if event.ObservationID == contributionID && event.NewStatus == "approved" && event.ModeratorID != nil && *event.ModeratorID == moderator.User.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("aucun evenement de moderation pour l'approbation")
	}

	_, problem, err = client.ApproveContribution(context.Background(), moderator.Tokens.AccessToken, contributionID, "")
	if err != nil {
		t.Fatalf("transport re-approbation : %v", err)
	}
	requireProblem(t, problem, http.StatusConflict, "not_modifiable", "re-approbation")
}

func TestModeratorRejectsFuelContribution(t *testing.T) {
	member := registerMember(t)
	contributionID := submitFuelContribution(t, member.Tokens.AccessToken)

	moderator := moderatorSession(t)
	detail, problem, err := client.RejectContribution(context.Background(), moderator.Tokens.AccessToken, contributionID, "Prix hors fourchette")
	if err != nil {
		t.Fatalf("transport rejet : %v", err)
	}
	requireNoProblem(t, problem, "rejet")
	if detail.Status != "rejected" {
		t.Fatalf("statut obtenu %s, attendu rejected", detail.Status)
	}

	contributions, problem, err := client.MyContributions(context.Background(), member.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("transport suivi : %v", err)
	}
	requireNoProblem(t, problem, "suivi")
	found := false
	for _, contribution := range contributions {
		if contribution.ID == contributionID && contribution.Status == "rejected" {
			found = true
		}
	}
	if !found {
		t.Fatal("rejet invisible dans le suivi du contributeur")
	}
}

func TestModeratorRejectionRequiresNote(t *testing.T) {
	member := registerMember(t)
	contributionID := submitFuelContribution(t, member.Tokens.AccessToken)
	moderator := moderatorSession(t)

	_, problem, err := client.RejectContribution(context.Background(), moderator.Tokens.AccessToken, contributionID, "")
	if err != nil {
		t.Fatalf("transport rejet sans motif : %v", err)
	}
	requireProblem(t, problem, http.StatusBadRequest, "invalid_moderation", "rejet sans motif")
}

func TestModerationQueueRequiresRoleAndAuth(t *testing.T) {
	member := registerMember(t)
	_, _, problem, err := client.ListModerationQueue(context.Background(), member.Tokens.AccessToken, "")
	if err != nil {
		t.Fatalf("transport membre : %v", err)
	}
	requireProblem(t, problem, http.StatusForbidden, "forbidden", "file par un membre")

	_, _, problem, err = client.ListModerationQueue(context.Background(), "", "")
	if err != nil {
		t.Fatalf("transport anonyme : %v", err)
	}
	requireProblem(t, problem, http.StatusUnauthorized, "unauthorized", "file sans jeton")
}

func TestQueueFiltersByStatus(t *testing.T) {
	admin := adminSession(t)
	productID := createProductAsAdmin(t, admin, uniqueBarcode())
	member := registerMember(t)
	contributionID := submitProductContribution(t, member.Tokens.AccessToken, productID)

	moderator := moderatorSession(t)
	detail, problem, err := client.ApproveContribution(context.Background(), moderator.Tokens.AccessToken, contributionID, "")
	if err != nil {
		t.Fatalf("transport approbation : %v", err)
	}
	requireNoProblem(t, problem, "approbation")
	if detail.Status != "approved" {
		t.Fatalf("statut obtenu %s", detail.Status)
	}

	pending, _, problem, err := client.ListModerationQueue(context.Background(), moderator.Tokens.AccessToken, "pending")
	if err != nil {
		t.Fatalf("transport file pending : %v", err)
	}
	requireNoProblem(t, problem, "file pending")
	for _, item := range pending {
		if item.ID == contributionID {
			t.Fatal("contribution approuvee encore dans la file pending")
		}
	}

	approved, _, problem, err := client.ListModerationQueue(context.Background(), moderator.Tokens.AccessToken, "approved")
	if err != nil {
		t.Fatalf("transport file approved : %v", err)
	}
	requireNoProblem(t, problem, "file approved")
	found := false
	for _, item := range approved {
		if item.ID == contributionID && item.Status == "approved" {
			found = true
		}
	}
	if !found {
		t.Fatal("contribution approuvee absente de la file approved")
	}
}

func TestModerationQueueRejectsUnknownStatus(t *testing.T) {
	moderator := moderatorSession(t)
	_, _, problem, err := client.ListModerationQueue(context.Background(), moderator.Tokens.AccessToken, "draft")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusBadRequest, "invalid_moderation", "statut inconnu")
}
