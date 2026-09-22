// Package api fournit un client HTTP type pour l'API publique OjoZone.
// Il decrit uniquement le contrat sur le fil (JSON) et n'importe aucun code
// du backend, afin de valider l'API comme le ferait un client externe.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client est un client HTTP minimal vers l'API OjoZone.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// New cree un client avec un timeout par defaut.
func New(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 10 * time.Second},
	}
}

// bytesReader adapte un corps brut en io.Reader.
func bytesReader(body []byte) *bytes.Reader {
	return bytes.NewReader(body)
}

// Result est l'enveloppe standard des reponses metier a succes.
type Result struct {
	Data json.RawMessage `json:"data"`
	Meta *PaginationMeta `json:"meta,omitempty"`
}

// DecodeData extrait le champ data d'une reponse a succes.
func (r *Result) DecodeData(target any) error {
	if len(r.Data) == 0 {
		return nil
	}
	return json.Unmarshal(r.Data, target)
}

// PaginationMeta est la pagination renvoyee par les listes.
type PaginationMeta struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

// Problem est la representation JSON de type application/problem+json
// renvoyee pour toutes les reponses d'erreur.
type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

// StatusCode retourne le code HTTP porte par le probleme.
func (p *Problem) StatusCode() int {
	if p == nil {
		return 0
	}
	return p.Status
}

// User est le profil expose par l'API.
type User struct {
	ID              string    `json:"id"`
	Email           string    `json:"email"`
	DisplayName     *string   `json:"display_name"`
	Role            string    `json:"role"`
	Locale          string    `json:"locale"`
	ReputationScore int       `json:"reputation_score"`
	EmailVerifiedAt *string   `json:"email_verified_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Tokens est la paire de jetons delivree par l'authentification.
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// AuthResult est le contenu de l'enveloppe auth.
type AuthResult struct {
	User   User   `json:"user"`
	Tokens Tokens `json:"tokens"`
}

// Contribution est une soumission de l'utilisateur courant.
type Contribution struct {
	ID            string  `json:"id"`
	Kind          string  `json:"kind"`
	Name          string  `json:"name"`
	Amount        string  `json:"amount"`
	Currency      string  `json:"currency"`
	Quantity      string  `json:"quantity"`
	UnitCode      string  `json:"unit_code"`
	ObservedAt    string  `json:"observed_at"`
	Status        string  `json:"status"`
	SourceID      string  `json:"source_id"`
	ModeratedAt   *string `json:"moderated_at"`
	ModeratorNote *string `json:"moderator_note"`
}

// Product est un produit du catalogue.
type Product struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Brand         *string         `json:"brand"`
	Barcode       *string         `json:"barcode"`
	ReferenceUnit string          `json:"reference_unit"`
	IsGeneric     bool            `json:"is_generic"`
	Attributes    json.RawMessage `json:"attributes"`
	CategorySlug  string          `json:"category_slug"`
	CreatedAt     time.Time       `json:"created_at"`
}

// ProductPrice est une observation de prix produit.
type ProductPrice struct {
	ID               string  `json:"id"`
	ProductID        string  `json:"product_id"`
	Amount           string  `json:"amount"`
	Currency         string  `json:"currency"`
	Quantity         string  `json:"quantity"`
	UnitCode         string  `json:"unit_code"`
	NormalizedAmount string  `json:"normalized_amount"`
	IsPromotion      bool    `json:"is_promotion"`
	ObservedAt       string  `json:"observed_at"`
	Status           string  `json:"status"`
	ConfidenceScore  *string `json:"confidence_score"`
	SourceRecordID   *string `json:"source_record_id"`
}

// send construit la requete HTTP et la transmet sans fermer le corps.
func (c *Client) send(ctx context.Context, method, path, token string, payload any) (*http.Response, error) {
	var reader io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("serialisation de la requete : %w", err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	return c.HTTP.Do(request)
}

// parseProblem decode un corps d'erreur en Problem avec le statut HTTP en secours.
func (c *Client) parseProblem(raw []byte, status int, path string) (*Problem, error) {
	var problem Problem
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &problem); err != nil {
			return nil, fmt.Errorf("decodage de l'erreur %s: %w", path, err)
		}
	}
	if problem.Status == 0 {
		problem.Status = status
	}
	if problem.Title == "" {
		problem.Title = http.StatusText(status)
	}
	if problem.Type == "" {
		problem.Type = http.StatusText(status)
	}
	return &problem, nil
}

// fetch s'attend a une reponse du schema {"data": ...} en cas de succes et
// decodera dataTarget ; toute reponse d'erreur est decodee en Problem avec le
// code HTTP en secours. Une reponse sans corps (204) est acceptee.
func (c *Client) fetch(ctx context.Context, method, path, token string, payload, dataTarget any) (*Problem, error) {
	response, err := c.send(ctx, method, path, token, payload)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		if len(raw) == 0 {
			return nil, nil
		}
		var wrapper Result
		if err := json.Unmarshal(raw, &wrapper); err != nil {
			return nil, fmt.Errorf("decodage de la reponse %s: %w", path, err)
		}
		if err := wrapper.DecodeData(dataTarget); err != nil {
			return nil, fmt.Errorf("decodage du champ data de %s: %w", path, err)
		}
		return nil, nil
	}
	return c.parseProblem(raw, response.StatusCode, path)
}
