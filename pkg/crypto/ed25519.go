// SPDX-License-Identifier: AGPL-3.0-or-later
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

	"github.com/btouchard/ackify-ce/internal/domain/models"
)

type Ed25519Signer struct {
    privateKey ed25519.PrivateKey
    publicKey  ed25519.PublicKey
}

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

func (s *Ed25519Signer) CreateSignature(docID string, user *models.User, timestamp time.Time, nonce string) (string, string, error) {
	payload := canonicalPayload(docID, user, timestamp, nonce)
	hash := sha256.Sum256(payload)
	signature := ed25519.Sign(s.privateKey, hash[:])

	return base64.StdEncoding.EncodeToString(hash[:]), base64.StdEncoding.EncodeToString(signature), nil
}

func (s *Ed25519Signer) GetPublicKey() string {
	return base64.StdEncoding.EncodeToString(s.publicKey)
}

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

func loadOrGenerateKeys() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	b64Key := strings.TrimSpace(os.Getenv("ACKIFY_ED25519_PRIVATE_KEY"))

	if b64Key != "" {
		keyBytes, err := base64.StdEncoding.DecodeString(b64Key)
		if err != nil || len(keyBytes) != ed25519.PrivateKeySize {
			return nil, nil, fmt.Errorf("invalid ACKIFY_ED25519_PRIVATE_KEY: %v", err)
		}

		privateKey := ed25519.PrivateKey(keyBytes)
		publicKey := privateKey.Public().(ed25519.PublicKey)

		return privateKey, publicKey, nil
	}

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate keys: %w", err)
	}

    // Do not print private keys. Warn about ephemeral key usage only.
    fmt.Println("[WARN] Generated ephemeral Ed25519 keypair. Set ACKIFY_ED25519_PRIVATE_KEY to persist across restarts.")

	return privateKey, publicKey, nil
}
