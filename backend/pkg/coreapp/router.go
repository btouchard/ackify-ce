// SPDX-License-Identifier: AGPL-3.0-or-later
package coreapp

import "github.com/go-chi/chi/v5"

type RouteRegistrar func(r chi.Router)

type HandlerGroups struct {
	RegisterPublic RouteRegistrar
	RegisterUser   RouteRegistrar
	RegisterAdmin  RouteRegistrar
}

func NewHandlerGroups(deps CoreDeps) HandlerGroups {
	h := NewCoreHandlers(deps)

	return HandlerGroups{
		RegisterPublic: func(r chi.Router) {
			r.Route("/documents", func(r chi.Router) {
				r.Get("/", h.Documents.HandleListDocuments)
				r.Get("/{docId}", h.Documents.HandleGetDocument)
				r.Get("/{docId}/signatures", h.Documents.HandleGetDocumentSignatures)
				r.Get("/{docId}/expected-signers", h.Documents.HandleGetExpectedSigners)
			})
		},

		RegisterUser: func(r chi.Router) {
			r.Post("/documents", h.Documents.HandleCreateDocument)
			r.Get("/documents/find-or-create", h.Documents.HandleFindOrCreateDocument)
			r.Get("/signatures", h.Signatures.HandleGetUserSignatures)
			r.Post("/signatures", h.Signatures.HandleCreateSignature)
			r.Get("/documents/{docId}/signatures/status", h.Signatures.HandleGetSignatureStatus)
		},

		RegisterAdmin: func(r chi.Router) {
			r.Route("/admin/documents", func(r chi.Router) {
				r.Get("/", h.Admin.HandleListDocuments)
				r.Get("/{docId}", h.Admin.HandleGetDocument)
				r.Get("/{docId}/signers", h.Admin.HandleGetDocumentWithSigners)
				r.Get("/{docId}/status", h.Admin.HandleGetDocumentStatus)
				r.Put("/{docId}/metadata", h.Admin.HandleUpdateDocumentMetadata)
				r.Delete("/{docId}", h.Admin.HandleDeleteDocument)
				r.Post("/{docId}/signers", h.Admin.HandleAddExpectedSigner)
				r.Delete("/{docId}/signers/{email}", h.Admin.HandleRemoveExpectedSigner)
				r.Post("/{docId}/signers/preview-csv", h.Admin.HandlePreviewCSV)
				r.Post("/{docId}/signers/import", h.Admin.HandleImportSigners)
				r.Post("/{docId}/reminders", h.Admin.HandleSendReminders)
				r.Get("/{docId}/reminders", h.Admin.HandleGetReminderHistory)
			})
		},
	}
}
