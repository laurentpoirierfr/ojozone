// Package auth gère les jetons d'accès, les jetons de rafraîchissement et le hachage des mots de passe.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ClaimNames documente les revendications personnalisées des jetons d'accès.
const (
	ClaimSessionID = "sid"
	ClaimRole      = "role"
)

// Claims contient les informations d'identité embarquées dans un jeton d'accès.
type Claims struct {
	SessionID string `json:"sid"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

// Manager émet et vérifie les jetons JWT HS256.
type Manager struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

// NewManager construit un gestionnaire de jetons.
// secret ne doit jamais être vide en production ; une longueur d'au moins 32 octets est recommandée.
func NewManager(secret, issuer string, accessTTL time.Duration) *Manager {
	return &Manager{secret: []byte(secret), issuer: issuer, accessTTL: accessTTL}
}

// AccessTTL retourne la durée de vie des jetons d'accès émis.
func (m *Manager) AccessTTL() time.Duration {
	return m.accessTTL
}

// IssueAccessToken signe un jeton d'accès court pour l'utilisateur et la session donnés.
func (m *Manager) IssueAccessToken(subject, role, sessionID string) (string, error) {
	now := time.Now()
	claims := Claims{
		SessionID: sessionID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// ParseAccessToken vérifie la signature, l'émetteur et la période de validité du jeton.
func (m *Manager) ParseAccessToken(token string) (Claims, error) {
	var claims Claims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithIssuer(m.issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil || !parsed.Valid {
		return Claims{}, errors.New("jeton d'accès invalide")
	}
	return claims, nil
}

// NewRefreshToken génère un jeton de rafraîchissement opaque et son empreinte SHA-256.
func (m *Manager) NewRefreshToken() (raw, hash string, err error) {
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(nonce)
	return raw, m.HashRefreshToken(raw), nil
}

// HashRefreshToken calcule l'empreinte de 64 caractères hexadécimaux d'un jeton de rafraîchissement.
func (m *Manager) HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// HashPassword hache un mot de passe avec bcrypt.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPassword vérifie un mot de passe contre son empreinte bcrypt.
func CheckPassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}
