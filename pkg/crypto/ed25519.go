package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"ackify/internal/domain/models"
)

// Ed25519Signer handles Ed25519 cryptographic operations
type Ed25519Signer struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewEd25519Signer creates a new Ed25519 signer
func NewEd25519Signer() (*Ed25519Signer, error) {
	privKey, pubKey, err := loadOrGenerateKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to load or generate keys: %w", err)
	}

	return &Ed25519Signer{
		privateKey: privKey,
		publicKey:  pubKey,
	}, nil
}

// CreateSignature creates a cryptographic signature for a document
func (s *Ed25519Signer) CreateSignature(docID string, user *models.User, timestamp time.Time, nonce string) (string, string, error) {
	payload := canonicalPayload(docID, user, timestamp, nonce)
	hash := sha256.Sum256(payload)
	signature := ed25519.Sign(s.privateKey, hash[:])

	return base64.StdEncoding.EncodeToString(hash[:]), base64.StdEncoding.EncodeToString(signature), nil
}

// GetPublicKey returns the base64 encoded public key
func (s *Ed25519Signer) GetPublicKey() string {
	return base64.StdEncoding.EncodeToString(s.publicKey)
}

// canonicalPayload creates a canonical payload for signing
func canonicalPayload(docID string, user *models.User, timestamp time.Time, nonce string) []byte {
	return []byte(fmt.Sprintf(
		"doc_id=%s\nuser_sub=%s\nuser_email=%s\nsigned_at=%s\nnonce=%s\n",
		docID,
		user.Sub,
		user.NormalizedEmail(),
		timestamp.UTC().Format(time.RFC3339Nano),
		nonce,
	))
}

// loadOrGenerateKeys loads existing keys or generates new ones
func loadOrGenerateKeys() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	b64Key := strings.TrimSpace(os.Getenv("ED25519_PRIVATE_KEY_B64"))

	if b64Key != "" {
		keyBytes, err := base64.StdEncoding.DecodeString(b64Key)
		if err != nil || len(keyBytes) != ed25519.PrivateKeySize {
			return nil, nil, fmt.Errorf("invalid ED25519_PRIVATE_KEY_B64: %v", err)
		}

		privateKey := ed25519.PrivateKey(keyBytes)
		publicKey := privateKey.Public().(ed25519.PublicKey)

		return privateKey, publicKey, nil
	}

	// Generate new keys
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	fmt.Printf("[WARN] Generated ephemeral Ed25519 keypair. Set ED25519_PRIVATE_KEY_B64 to persist: %s\n",
		base64.StdEncoding.EncodeToString(privateKey))

	return privateKey, publicKey, nil
}
