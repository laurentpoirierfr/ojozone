package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Register cree un compte membre.
func (c *Client) Register(ctx context.Context, email, password, locale string) (AuthResult, *Problem, error) {
	var result AuthResult
	problem, err := c.fetch(ctx, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"email": email, "password": password, "locale": locale,
	}, &result)
	return result, problem, err
}

// Login ouvre une session.
func (c *Client) Login(ctx context.Context, email, password string) (AuthResult, *Problem, error) {
	var result AuthResult
	problem, err := c.fetch(ctx, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"email": email, "password": password,
	}, &result)
	return result, problem, err
}

// Refresh renouvelle les jetons par rotation.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (AuthResult, *Problem, error) {
	var result AuthResult
	problem, err := c.fetch(ctx, http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{
		"refresh_token": refreshToken,
	}, &result)
	return result, problem, err
}

// Logout reveque la session du jeton de rafraichissement.
func (c *Client) Logout(ctx context.Context, refreshToken string) (*Problem, error) {
	return c.fetch(ctx, http.MethodPost, "/api/v1/auth/logout", "", map[string]string{
		"refresh_token": refreshToken,
	}, nil)
}

// Me renvoie le profil courant.
func (c *Client) Me(ctx context.Context, accessToken string) (User, *Problem, error) {
	var user User
	problem, err := c.fetch(ctx, http.MethodGet, "/api/v1/me", accessToken, nil, &user)
	return user, problem, err
}

// PatchMe modifie partiellement le profil courant.
func (c *Client) PatchMe(ctx context.Context, accessToken string, fields map[string]any) (User, *Problem, error) {
	var user User
	problem, err := c.fetch(ctx, http.MethodPatch, "/api/v1/me", accessToken, fields, &user)
	return user, problem, err
}

// DeleteMe supprime le compte courant (204).
func (c *Client) DeleteMe(ctx context.Context, accessToken string) (*Problem, error) {
	return c.fetch(ctx, http.MethodDelete, "/api/v1/me", accessToken, nil, nil)
}

// MyContributions liste les contributions de l'utilisateur courant.
func (c *Client) MyContributions(ctx context.Context, accessToken string) ([]Contribution, *Problem, error) {
	var contributions []Contribution
	problem, err := c.fetch(ctx, http.MethodGet, "/api/v1/me/contributions", accessToken, nil, &contributions)
	return contributions, problem, err
}

// SubmitProductContribution propose un prix produit (202 et statut pending).
func (c *Client) SubmitProductContribution(ctx context.Context, accessToken string, payload map[string]any) (ContributionResult, *Problem, error) {
	var result ContributionResult
	problem, err := c.fetch(ctx, http.MethodPost, "/api/v1/contributions/product-prices", accessToken, payload, &result)
	return result, problem, err
}

// SubmitFuelContribution propose un prix carburant (202 et statut pending).
func (c *Client) SubmitFuelContribution(ctx context.Context, accessToken string, payload map[string]any) (ContributionResult, *Problem, error) {
	var result ContributionResult
	problem, err := c.fetch(ctx, http.MethodPost, "/api/v1/contributions/fuel-prices", accessToken, payload, &result)
	return result, problem, err
}

// GetContribution expose une contribution au proprietaire ou a la moderation.
func (c *Client) GetContribution(ctx context.Context, accessToken, id string) (ContributionDetail, *Problem, error) {
	var detail ContributionDetail
	problem, err := c.fetch(ctx, http.MethodGet, "/api/v1/contributions/"+id, accessToken, nil, &detail)
	return detail, problem, err
}

// PatchContribution corrige une contribution encore pending.
func (c *Client) PatchContribution(ctx context.Context, accessToken, id string, fields map[string]any) (ContributionDetail, *Problem, error) {
	var detail ContributionDetail
	problem, err := c.fetch(ctx, http.MethodPatch, "/api/v1/contributions/"+id, accessToken, fields, &detail)
	return detail, problem, err
}

// DeleteContribution retire une contribution encore pending (204).
func (c *Client) DeleteContribution(ctx context.Context, accessToken, id string) (*Problem, error) {
	return c.fetch(ctx, http.MethodDelete, "/api/v1/contributions/"+id, accessToken, nil, nil)
}

// ListProducts recherche les produits publiquement (filtre search optionnel).
func (c *Client) ListProducts(ctx context.Context, search string) ([]Product, PaginationMeta, *Problem, error) {
	path := "/api/v1/products"
	if search != "" {
		path += "?q=" + url.QueryEscape(search)
	}
	var products []Product
	meta, problem, err := c.fetchList(ctx, http.MethodGet, path, "", nil, &products)
	return products, meta, problem, err
}

// CreateProduct cree ou met a jour un produit (upsert, 200).
func (c *Client) CreateProduct(ctx context.Context, accessToken string, payload map[string]any) (Product, *Problem, error) {
	var product Product
	problem, err := c.fetch(ctx, http.MethodPost, "/api/v1/products", accessToken, payload, &product)
	return product, problem, err
}

// GetProduct renvoie un produit.
func (c *Client) GetProduct(ctx context.Context, id string) (Product, *Problem, error) {
	var product Product
	problem, err := c.fetch(ctx, http.MethodGet, "/api/v1/products/"+id, "", nil, &product)
	return product, problem, err
}

// ListProductPrices renvoie les prix d'un produit.
func (c *Client) ListProductPrices(ctx context.Context, productID string) ([]ProductPrice, PaginationMeta, *Problem, error) {
	var prices []ProductPrice
	meta, problem, err := c.fetchList(ctx, http.MethodGet, "/api/v1/products/"+productID+"/prices", "", nil, &prices)
	return prices, meta, problem, err
}

// CreateProductPrice cree ou met a jour une observation de prix (upsert, 200).
func (c *Client) CreateProductPrice(ctx context.Context, accessToken string, payload map[string]any) (ProductPrice, *Problem, error) {
	var price ProductPrice
	problem, err := c.fetch(ctx, http.MethodPost, "/api/v1/product-prices", accessToken, payload, &price)
	return price, problem, err
}

// GetProductPrice renvoie une observation de prix.
func (c *Client) GetProductPrice(ctx context.Context, id string) (ProductPrice, *Problem, error) {
	var price ProductPrice
	problem, err := c.fetch(ctx, http.MethodGet, "/api/v1/product-prices/"+id, "", nil, &price)
	return price, problem, err
}

// ListAdminUsers renvoie la liste des utilisateurs (route administrateur).
func (c *Client) ListAdminUsers(ctx context.Context, accessToken string) ([]User, *Problem, error) {
	var users []User
	problem, err := c.fetch(ctx, http.MethodGet, "/api/v1/admin/users", accessToken, nil, &users)
	return users, problem, err
}

// Raw envoie un corps brut (non serialise) pour tester les corps invalides.
// Le corps de la reponse reste ouvert ; l'appelant doit le fermer.
func (c *Client) Raw(ctx context.Context, method, path, token string, rawBody []byte) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytesReader(rawBody))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	return c.HTTP.Do(request)
}

// fetchList comme fetch mais expose egalement la pagination de l'enveloppe.
func (c *Client) fetchList(ctx context.Context, method, path, token string, payload any, dataTarget any) (PaginationMeta, *Problem, error) {
	response, err := c.send(ctx, method, path, token, payload)
	if err != nil {
		return PaginationMeta{}, nil, err
	}
	defer response.Body.Close()

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return PaginationMeta{}, nil, err
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		var wrapper Result
		if err := json.Unmarshal(raw, &wrapper); err != nil {
			return PaginationMeta{}, nil, fmt.Errorf("decodage de la reponse %s: %w", path, err)
		}
		if err := wrapper.DecodeData(dataTarget); err != nil {
			return PaginationMeta{}, nil, fmt.Errorf("decodage du champ data de %s: %w", path, err)
		}
		meta := PaginationMeta{}
		if wrapper.Meta != nil {
			meta = *wrapper.Meta
		}
		return meta, nil, nil
	}
	problem, err := c.parseProblem(raw, response.StatusCode, path)
	return PaginationMeta{}, problem, err
}
