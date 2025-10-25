// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "time"

// OAuthSession represents an OAuth session with encrypted refresh token
type OAuthSession struct {
	ID                    int64
	SessionID             string
	UserSub               string
	RefreshTokenEncrypted []byte
	AccessTokenExpiresAt  time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
	LastRefreshedAt       *time.Time
	UserAgent             string
	IPAddress             string
}
