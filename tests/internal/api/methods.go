package api

import (
	"context"
	"net/http"
	"net/url"
)

// ListProducts recherche les produits publiquement (filtre search optionnel).
func (c *Client) ListProducts(ctx context.Context, search string) ([]Product, PaginationMeta, *Problem, error) {
	path := "/api/v1/products"
	if search != "" {
		path += "?search=" + url.QueryEscape(search)
	}
	var wrapper Result
	response, err := c.Do(ctx, http.MethodGet, path, "", nil, &wrapper)
	if err != nil {
		return nil, PaginationMeta{}, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return nil, PaginationMeta{}, problem, err
	}
	var products []Product
	if err := wrapper.DecodeData(&products); err != nil {
		return nil, PaginationMeta{}, nil, err
	}
	meta := PaginationMeta{}
	if wrapper.Meta != nil {
		meta = *wrapper.Meta
	}
	return products, meta, nil, nil
}

// Raw envoie un corps brut (non serialise) pour tester les corps invalides.
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

// ListProductPrices renvoie les prix d'un produit.
func (c *Client) ListProductPrices(ctx context.Context, productID string) ([]ProductPrice, PaginationMeta, *Problem, error) {
	var wrapper Result
	response, err := c.Do(ctx, http.MethodGet, "/api/v1/products/"+productID+"/prices", "", nil, &wrapper)
	if err != nil {
		return nil, PaginationMeta{}, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return nil, PaginationMeta{}, problem, err
	}
	var prices []ProductPrice
	if err := wrapper.DecodeData(&prices); err != nil {
		return nil, PaginationMeta{}, nil, err
	}
	meta := PaginationMeta{}
	if wrapper.Meta != nil {
		meta = *wrapper.Meta
	}
	return prices, meta, nil, nil
}

// ListAdminUsers renvoie la liste des utilisateurs (route administrateur).
func (c *Client) ListAdminUsers(ctx context.Context, accessToken string) ([]User, *Problem, error) {
	var wrapper Result
	response, err := c.Do(ctx, http.MethodGet, "/api/v1/admin/users", accessToken, nil, &wrapper)
	if err != nil {
		return nil, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return nil, problem, err
	}
	var users []User
	if err := wrapper.DecodeData(&users); err != nil {
		return nil, nil, err
	}
	return users, nil, nil
}

// Register cree un compte membre.
func (c *Client) Register(ctx context.Context, email, password, locale string) (AuthResult, *Problem, error) {
	body := map[string]string{"email": email, "password": password, "locale": locale}
	var wrapper Result
	response, err := c.Do(ctx, http.MethodPost, "/api/v1/auth/register", "", body, &wrapper)
	if err != nil {
		return AuthResult{}, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return AuthResult{}, problem, err
	}
	var result AuthResult
	if err := wrapper.DecodeData(&result); err != nil {
		return AuthResult{}, nil, err
	}
	return result, nil, nil
}

// Login ouvre une session.
func (c *Client) Login(ctx context.Context, email, password string) (AuthResult, *Problem, error) {
	body := map[string]string{"email": email, "password": password}
	var wrapper Result
	response, err := c.Do(ctx, http.MethodPost, "/api/v1/auth/login", "", body, &wrapper)
	if err != nil {
		return AuthResult{}, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return AuthResult{}, problem, err
	}
	var result AuthResult
	if err := wrapper.DecodeData(&result); err != nil {
		return AuthResult{}, nil, err
	}
	return result, nil, nil
}

// Refresh renouvelle les jetons par rotation.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (AuthResult, *Problem, error) {
	body := map[string]string{"refresh_token": refreshToken}
	var wrapper Result
	response, err := c.Do(ctx, http.MethodPost, "/api/v1/auth/refresh", "", body, &wrapper)
	if err != nil {
		return AuthResult{}, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return AuthResult{}, problem, err
	}
	var result AuthResult
	if err := wrapper.DecodeData(&result); err != nil {
		return AuthResult{}, nil, err
	}
	return result, nil, nil
}

// Logout reveque la session du jeton de rafraichissement.
func (c *Client) Logout(ctx context.Context, refreshToken string) (*Problem, error) {
	body := map[string]string{"refresh_token": refreshToken}
	var wrapper Result
	response, err := c.Do(ctx, http.MethodPost, "/api/v1/auth/logout", "", body, &wrapper)
	if err != nil {
		return nil, err
	}
	return decodeOutcome(response, &wrapper)
}

// Me renvoie le profil courant.
func (c *Client) Me(ctx context.Context, accessToken string) (User, *Problem, error) {
	var wrapper Result
	response, err := c.Do(ctx, http.MethodGet, "/api/v1/me", accessToken, nil, &wrapper)
	if err != nil {
		return User{}, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return User{}, problem, err
	}
	var user User
	if err := wrapper.DecodeData(&user); err != nil {
		return User{}, nil, err
	}
	return user, nil, nil
}

// PatchMe modifie partiellement le profil courant.
func (c *Client) PatchMe(ctx context.Context, accessToken string, fields map[string]any) (User, *Problem, error) {
	var wrapper Result
	response, err := c.Do(ctx, http.MethodPatch, "/api/v1/me", accessToken, fields, &wrapper)
	if err != nil {
		return User{}, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return User{}, problem, err
	}
	var user User
	if err := wrapper.DecodeData(&user); err != nil {
		return User{}, nil, err
	}
	return user, nil, nil
}

// DeleteMe supprime le compte courant (204).
func (c *Client) DeleteMe(ctx context.Context, accessToken string) (*Problem, error) {
	var wrapper Result
	response, err := c.Do(ctx, http.MethodDelete, "/api/v1/me", accessToken, nil, &wrapper)
	if err != nil {
		return nil, err
	}
	return decodeOutcome(response, &wrapper)
}

// MyContributions liste les contributions de l'utilisateur courant.
func (c *Client) MyContributions(ctx context.Context, accessToken string) ([]Contribution, *Problem, error) {
	var wrapper Result
	response, err := c.Do(ctx, http.MethodGet, "/api/v1/me/contributions", accessToken, nil, &wrapper)
	if err != nil {
		return nil, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return nil, problem, err
	}
	var contributions []Contribution
	if err := wrapper.DecodeData(&contributions); err != nil {
		return nil, nil, err
	}
	return contributions, nil, nil
}

// CreateProduct cree ou met a jour un produit (upsert, 200).
func (c *Client) CreateProduct(ctx context.Context, accessToken string, payload map[string]any) (Product, *Problem, error) {
	var wrapper Result
	response, err := c.Do(ctx, http.MethodPost, "/api/v1/products", accessToken, payload, &wrapper)
	if err != nil {
		return Product{}, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return Product{}, problem, err
	}
	var product Product
	if err := wrapper.DecodeData(&product); err != nil {
		return Product{}, nil, err
	}
	return product, nil, nil
}

// GetProduct renvoie un produit.
func (c *Client) GetProduct(ctx context.Context, id string) (Product, *Problem, error) {
	var wrapper Result
	response, err := c.Do(ctx, http.MethodGet, "/api/v1/products/"+id, "", nil, &wrapper)
	if err != nil {
		return Product{}, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return Product{}, problem, err
	}
	var product Product
	if err := wrapper.DecodeData(&product); err != nil {
		return Product{}, nil, err
	}
	return product, nil, nil
}

// CreateProductPrice cree ou met a jour une observation de prix (upsert, 200).
func (c *Client) CreateProductPrice(ctx context.Context, accessToken string, payload map[string]any) (ProductPrice, *Problem, error) {
	var wrapper Result
	response, err := c.Do(ctx, http.MethodPost, "/api/v1/product-prices", accessToken, payload, &wrapper)
	if err != nil {
		return ProductPrice{}, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return ProductPrice{}, problem, err
	}
	var price ProductPrice
	if err := wrapper.DecodeData(&price); err != nil {
		return ProductPrice{}, nil, err
	}
	return price, nil, nil
}

// GetProductPrice renvoie une observation de prix.
func (c *Client) GetProductPrice(ctx context.Context, id string) (ProductPrice, *Problem, error) {
	var wrapper Result
	response, err := c.Do(ctx, http.MethodGet, "/api/v1/product-prices/"+id, "", nil, &wrapper)
	if err != nil {
		return ProductPrice{}, nil, err
	}
	if problem, err := decodeOutcome(response, &wrapper); problem != nil || err != nil {
		return ProductPrice{}, problem, err
	}
	var price ProductPrice
	if err := wrapper.DecodeData(&price); err != nil {
		return ProductPrice{}, nil, err
	}
	return price, nil, nil
}
