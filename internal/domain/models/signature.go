// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/btouchard/ackify-ce/pkg/services"
)

type Signature struct {
	ID          int64     `json:"id" db:"id"`
	DocID       string    `json:"doc_id" db:"doc_id"`
	UserSub     string    `json:"user_sub" db:"user_sub"`
	UserEmail   string    `json:"user_email" db:"user_email"`
	UserName    string    `json:"user_name,omitempty" db:"user_name"`
	SignedAtUTC time.Time `json:"signed_at" db:"signed_at"`
	PayloadHash string    `json:"payload_hash" db:"payload_hash"`
	Signature   string    `json:"signature" db:"signature"`
	Nonce       string    `json:"nonce" db:"nonce"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	Referer     *string   `json:"referer,omitempty" db:"referer"`
	PrevHash    *string   `json:"prev_hash,omitempty" db:"prev_hash"`
}

func (s *Signature) GetServiceInfo() *services.ServiceInfo {
	if s.Referer == nil {
		return nil
	}
	return services.DetectServiceFromReferrer(*s.Referer)
}

type SignatureRequest struct {
	DocID   string
	User    *User
	Referer *string
}

type SignatureStatus struct {
	DocID     string
	UserEmail string
	IsSigned  bool
	SignedAt  *time.Time
}

// ComputeRecordHash Stable record hash supports tamper-evident chaining and integrity checks across migrations.
func (s *Signature) ComputeRecordHash() string {
	data := fmt.Sprintf("%d|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s",
		s.ID,
		s.DocID,
		s.UserSub,
		s.UserEmail,
		s.UserName,
		s.SignedAtUTC.Format(time.RFC3339Nano),
		s.PayloadHash,
		s.Signature,
		s.Nonce,
		s.CreatedAt.Format(time.RFC3339Nano),
		func() string {
			if s.Referer != nil {
				return *s.Referer
			}
			return ""
		}(),
	)

	hash := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:])
}
