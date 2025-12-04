// SPDX-License-Identifier: AGPL-3.0-or-later
package authorization

import (
	"context"
	"strings"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
)

type CEDocumentAuthorizer struct {
	adminEmails        []string
	onlyAdminCanCreate bool
}

func NewCEDocumentAuthorizer(adminEmails []string, onlyAdminCanCreate bool) *CEDocumentAuthorizer {
	return &CEDocumentAuthorizer{
		adminEmails:        adminEmails,
		onlyAdminCanCreate: onlyAdminCanCreate,
	}
}

func (a *CEDocumentAuthorizer) CanCreateDocument(ctx context.Context, user *models.User) bool {
	if !a.onlyAdminCanCreate {
		return true
	}

	if user == nil {
		return false
	}

	for _, adminEmail := range a.adminEmails {
		if strings.EqualFold(user.Email, adminEmail) {
			return true
		}
	}

	return false
}
