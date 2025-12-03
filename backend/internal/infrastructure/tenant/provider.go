// SPDX-License-Identifier: AGPL-3.0-or-later
package tenant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Provider defines the interface for obtaining the current tenant ID.
type Provider interface {
	CurrentTenant(ctx context.Context) (uuid.UUID, error)
}

// SingleTenantProvider implements Provider for single-tenant.
// It caches the tenant ID from instance_metadata at startup.
type SingleTenantProvider struct {
	id uuid.UUID
}

// CurrentTenant returns the cached instance tenant ID.
func (p *SingleTenantProvider) CurrentTenant(_ context.Context) (uuid.UUID, error) {
	return p.id, nil
}

// NewSingleTenantProviderWithContext is like NewSingleTenantProvider but accepts a context.
func NewSingleTenantProviderWithContext(ctx context.Context, db *sql.DB) (*SingleTenantProvider, error) {
	var id uuid.UUID
	err := db.QueryRowContext(ctx, `SELECT id FROM instance_metadata LIMIT 1`).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("instance_metadata table is empty - migrations may not have run correctly")
		}
		return nil, fmt.Errorf("failed to read tenant ID from instance_metadata: %w", err)
	}

	return &SingleTenantProvider{id: id}, nil
}
