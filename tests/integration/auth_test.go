package integration

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestRegisterCreatesMemberAccount(t *testing.T) {
	result, problem, err := client.Register(context.Background(), uniqueEmail("register"), "mot-de-passe-123", "fr")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireNoProblem(t, problem, "inscription")
	if result.User.Role != "member" {
		t.Fatalf("role inattendu : %q", result.User.Role)
	}
	if result.User.Email == "" || result.User.ID == "" {
		t.Fatal("profil sans identifiant ou email")
	}
	if result.Tokens.TokenType != "Bearer" {
		t.Fatalf("type de jeton inattendu : %q", result.Tokens.TokenType)
	}
}

func TestRegisterNormalizesEmailToLowercase(t *testing.T) {
	email := uniqueEmail("CaSe")
	result, problem, err := client.Register(context.Background(), email, "mot-de-passe-123", "fr")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireNoProblem(t, problem, "inscription")
	if result.User.Email != strings.ToLower(email) {
		t.Fatalf("email non normalise : %q, attendu %q", result.User.Email, strings.ToLower(email))
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	email := uniqueEmail("dup")
	if _, problem, err := client.Register(context.Background(), email, "mot-de-passe-123", "fr"); err != nil {
		t.Fatalf("transport : %v", err)
	} else {
		requireNoProblem(t, problem, "premiere inscription")
	}
	_, problem, err := client.Register(context.Background(), email, "autre-mot-de-passe", "fr")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusConflict, "email_already_used", "doublon")
}

func TestRegisterValidatesFields(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		locale   string
	}{
		{name: "email invalide", email: "nope", password: "mot-de-passe-123", locale: "fr"},
		{name: "mot de passe court", email: uniqueEmail("pwd"), password: "court", locale: "fr"},
		{name: "locale inconnue", email: uniqueEmail("loc"), password: "mot-de-passe-123", locale: "de"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, problem, err := client.Register(context.Background(), test.email, test.password, test.locale)
			if err != nil {
				t.Fatalf("transport : %v", err)
			}
			requireProblem(t, problem, http.StatusBadRequest, "", "validation")
		})
	}
}

func TestLoginOpensSessionAndReturnsProfile(t *testing.T) {
	email := uniqueEmail("login")
	registered, problem, err := client.Register(context.Background(), email, "mot-de-passe-123", "fr")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireNoProblem(t, problem, "inscription")

	result, problem, err := client.Login(context.Background(), email, "mot-de-passe-123")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireNoProblem(t, problem, "connexion")
	if result.User.ID != registered.User.ID {
		t.Fatalf("profil different de l'inscription : %s != %s", result.User.ID, registered.User.ID)
	}
	if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" {
		t.Fatal("connexion sans jetons")
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	email := uniqueEmail("badpwd")
	if _, problem, err := client.Register(context.Background(), email, "mot-de-passe-123", "fr"); err != nil {
		t.Fatalf("transport : %v", err)
	} else {
		requireNoProblem(t, problem, "inscription")
	}
	_, problem, err := client.Login(context.Background(), email, "mauvais-mot-de-passe")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusUnauthorized, "invalid_credentials", "mot de passe")
}

func TestLoginRejectsUnknownEmail(t *testing.T) {
	_, problem, err := client.Login(context.Background(), uniqueEmail("absent"), "mot-de-passe-123")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusUnauthorized, "invalid_credentials", "email inconnu")
}

func TestMeRequiresAndReturnsProfile(t *testing.T) {
	registered := registerMember(t)
	user, problem, err := client.Me(context.Background(), registered.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireNoProblem(t, problem, "lecture du profil")
	if user.ID != registered.User.ID || user.Email == "" {
		t.Fatalf("profil inattendu : %+v", user)
	}
}

func TestMeRejectsMissingOrInvalidToken(t *testing.T) {
	_, problem, err := client.Me(context.Background(), "")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusUnauthorized, "unauthorized", "sans jeton")

	_, problem, err = client.Me(context.Background(), "pas-un-jwt")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusUnauthorized, "invalid_token", "jeton invalide")
}

func TestRefreshRotatesTokenAndRejectsReuse(t *testing.T) {
	registered := registerMember(t)
	result, problem, err := client.Refresh(context.Background(), registered.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireNoProblem(t, problem, "renouvellement")
	if result.Tokens.RefreshToken == registered.Tokens.RefreshToken {
		t.Fatal("le jeton de rafraichissement doit etre remplace")
	}
	if result.Tokens.AccessToken == "" {
		t.Fatal("renouvellement sans jeton d'acces")
	}

	_, problem, err = client.Refresh(context.Background(), registered.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusUnauthorized, "unauthorized", "rejeu")
}

func TestLogoutRevokesSession(t *testing.T) {
	registered := registerMember(t)
	problem, err := client.Logout(context.Background(), registered.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireNoProblem(t, problem, "deconnexion")

	_, problem, err = client.Refresh(context.Background(), registered.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusUnauthorized, "unauthorized", "session revequee")
}

func TestPatchMeUpdatesOnlyProvidedFields(t *testing.T) {
	registered := registerMember(t)
	displayName := "Marie Dupont"
	user, problem, err := client.PatchMe(context.Background(), registered.Tokens.AccessToken, map[string]any{
		"display_name": displayName,
		"locale":       "en",
	})
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireNoProblem(t, problem, "mise a jour profil")
	if user.DisplayName == nil || *user.DisplayName != displayName {
		t.Fatalf("display_name non applique : %+v", user.DisplayName)
	}
	if user.Locale != "en" {
		t.Fatalf("locale non appliquee : %q", user.Locale)
	}
	if user.Email == "" || user.Role != "member" {
		t.Fatalf("champs protégés modifiés : %+v", user)
	}
}

func TestPatchMeRejectsInvalidLocale(t *testing.T) {
	registered := registerMember(t)
	_, problem, err := client.PatchMe(context.Background(), registered.Tokens.AccessToken, map[string]any{"locale": "de"})
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusBadRequest, "invalid_profile", "locale invalide")
}

func TestDeleteMeAnonymizesAndRevokesAccess(t *testing.T) {
	registered := registerMember(t)
	problem, err := client.DeleteMe(context.Background(), registered.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireNoProblem(t, problem, "suppression")

	_, problem, err = client.Refresh(context.Background(), registered.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireProblem(t, problem, http.StatusUnauthorized, "unauthorized", "session supprimee")

	_, problem, err = client.Login(context.Background(), registered.User.Email, "mot-de-passe-123")
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	if problem == nil {
		t.Fatal("la connexion doit échouer apres anonymisation")
	}
}

func TestMyContributionsIsEmptyForNewMember(t *testing.T) {
	registered := registerMember(t)
	contributions, problem, err := client.MyContributions(context.Background(), registered.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("transport : %v", err)
	}
	requireNoProblem(t, problem, "contributions")
	if len(contributions) != 0 {
		t.Fatalf("contributions inattendues pour un nouveau compte : %+v", contributions)
	}
}
