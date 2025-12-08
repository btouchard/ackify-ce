// SPDX-License-Identifier: AGPL-3.0-or-later
package coreapp

import (
	"github.com/btouchard/ackify-ce/internal/presentation/api/admin"
	"github.com/btouchard/ackify-ce/internal/presentation/api/documents"
	"github.com/btouchard/ackify-ce/internal/presentation/api/signatures"
)

type CoreHandlers struct {
	Documents  *documents.Handler
	Signatures *signatures.Handler
	Admin      *admin.Handler
}

func NewCoreHandlers(deps CoreDeps) *CoreHandlers {
	return &CoreHandlers{
		Documents: documents.NewHandler(
			deps.Signatures,
			deps.Documents,
			deps.Documents,       // DocumentService implements documentRepository interface
			deps.ExpectedSigners, // ExpectedSignerService implements expectedSignerRepository interface
			deps.WebhookPublisher,
			deps.DocumentAuthorizer,
		),
		Signatures: signatures.NewHandler(
			deps.Signatures,
			deps.ExpectedSigners, // ExpectedSignerService implements expectedSignerStatsRepo interface
			deps.WebhookPublisher,
		),
		Admin: admin.NewHandler(
			deps.Documents,       // DocumentService implements documentRepository interface
			deps.ExpectedSigners, // ExpectedSignerService implements expectedSignerRepository interface
			deps.Reminders,
			deps.Signatures,
			deps.BaseURL,
			deps.ImportMaxSigners,
		),
	}
}
