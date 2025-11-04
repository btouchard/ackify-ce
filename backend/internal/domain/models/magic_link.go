// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "time"

// MagicLinkToken représente un token de connexion Magic Link
type MagicLinkToken struct {
	ID                 int64      `json:"id" db:"id"`
	Token              string     `json:"token" db:"token"`
	Email              string     `json:"email" db:"email"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt          time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt             *time.Time `json:"used_at,omitempty" db:"used_at"`
	UsedByIP           *string    `json:"used_by_ip,omitempty" db:"used_by_ip"`
	UsedByUserAgent    *string    `json:"used_by_user_agent,omitempty" db:"used_by_user_agent"`
	RedirectTo         string     `json:"redirect_to" db:"redirect_to"` // URL destination après auth (ex: /?doc=xxx)
	CreatedByIP        string     `json:"created_by_ip" db:"created_by_ip"`
	CreatedByUserAgent string     `json:"created_by_user_agent" db:"created_by_user_agent"`
}

// IsValid vérifie si le token est valide (non expiré, non utilisé)
func (t *MagicLinkToken) IsValid() bool {
	if t.UsedAt != nil {
		return false // Déjà utilisé
	}
	if time.Now().After(t.ExpiresAt) {
		return false // Expiré
	}
	return true
}

// MagicLinkAuthAttempt représente une tentative d'authentification
type MagicLinkAuthAttempt struct {
	ID            int64     `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	Success       bool      `json:"success" db:"success"`
	FailureReason string    `json:"failure_reason,omitempty" db:"failure_reason"`
	IPAddress     string    `json:"ip_address" db:"ip_address"`
	UserAgent     string    `json:"user_agent,omitempty" db:"user_agent"`
	AttemptedAt   time.Time `json:"attempted_at" db:"attempted_at"`
}
