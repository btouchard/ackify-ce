package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/internal/presentation/admin"
	"github.com/btouchard/ackify-ce/pkg/services"
)

type signatureService interface {
	CreateSignature(ctx context.Context, request *models.SignatureRequest) error
	GetSignatureStatus(ctx context.Context, docID string, user *models.User) (*models.SignatureStatus, error)
	GetSignatureByDocAndUser(ctx context.Context, docID string, user *models.User) (*models.Signature, error)
	GetDocumentSignatures(ctx context.Context, docID string) ([]*models.Signature, error)
	GetUserSignatures(ctx context.Context, user *models.User) ([]*models.Signature, error)
	CheckUserSignature(ctx context.Context, docID, userIdentifier string) (bool, error)
}

// SignatureHandlers handles signature-related HTTP requests
type SignatureHandlers struct {
	signatureService signatureService
	userService      userService
	template         *template.Template
	baseURL          string
	organisation     string
}

// NewSignatureHandlers creates new signature handlers
func NewSignatureHandlers(signatureService signatureService, userService userService, tmpl *template.Template, baseURL, organisation string) *SignatureHandlers {
	return &SignatureHandlers{
		signatureService: signatureService,
		userService:      userService,
		template:         tmpl,
		baseURL:          baseURL,
		organisation:     organisation,
	}
}

// PageData represents data passed to templates
type PageData struct {
	User         *models.User
	Organisation string
	Year         int
	DocID        string
	Already      bool
	SignedAt     string
	TemplateName string
	BaseURL      string
	Signatures   []*models.Signature
	IsAdmin      bool
	ServiceInfo  *struct {
		Name     string
		Icon     string
		Type     string
		Referrer string
	}
}

// HandleIndex serves the main index page
func (h *SignatureHandlers) HandleIndex(w http.ResponseWriter, r *http.Request) {
	user, _ := h.userService.GetUser(r)
	h.render(w, r, "index", PageData{User: user, Organisation: h.organisation})
}

// HandleSignGET displays the signature page
func (h *SignatureHandlers) HandleSignGET(w http.ResponseWriter, r *http.Request) {
	user, err := h.userService.GetUser(r)
	if err != nil {
		HandleError(w, err)
		return
	}

	docID, err := validateDocID(r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx := r.Context()
	status, err := h.signatureService.GetSignatureStatus(ctx, docID, user)
	if err != nil {
		HandleError(w, err)
		return
	}

	signedAt := ""
	var serviceInfo *struct {
		Name     string
		Icon     string
		Type     string
		Referrer string
	}

	// First try to get service info from URL parameter (always present when coming from embed)
	if referrerParam := r.URL.Query().Get("referrer"); referrerParam != "" {
		if sigServiceInfo := services.DetectServiceFromReferrer(referrerParam); sigServiceInfo != nil {
			serviceInfo = &struct {
				Name     string
				Icon     string
				Type     string
				Referrer string
			}{
				Name:     sigServiceInfo.Name,
				Icon:     sigServiceInfo.Icon,
				Type:     sigServiceInfo.Type,
				Referrer: sigServiceInfo.Referrer,
			}
		}
	}

	if status.IsSigned {
		// Get full signature to access referer information
		signature, err := h.signatureService.GetSignatureByDocAndUser(ctx, docID, user)
		if err == nil && signature != nil {
			if signature.SignedAtUTC.IsZero() == false {
				signedAt = signature.SignedAtUTC.Format("02/01/2006 à 15:04:05")
			}

			if serviceInfo == nil && signature.Referer != nil {
				if sigServiceInfo := signature.GetServiceInfo(); sigServiceInfo != nil {
					serviceInfo = &struct {
						Name     string
						Icon     string
						Type     string
						Referrer string
					}{
						Name:     sigServiceInfo.Name,
						Icon:     sigServiceInfo.Icon,
						Type:     sigServiceInfo.Type,
						Referrer: sigServiceInfo.Referrer,
					}
				}
			}
		}
	}

	if signedAt == "" && status.SignedAt != nil {
		signedAt = status.SignedAt.Format("02/01/2006 à 15:04:05")
	}

	h.render(w, r, "sign", PageData{
		User:        user,
		DocID:       docID,
		Already:     status.IsSigned,
		SignedAt:    signedAt,
		BaseURL:     h.baseURL,
		ServiceInfo: serviceInfo,
	})
}

// HandleSignPOST processes signature creation
func (h *SignatureHandlers) HandleSignPOST(w http.ResponseWriter, r *http.Request) {
	user, err := h.userService.GetUser(r)
	if err != nil {
		if docID := r.FormValue("doc"); docID != "" {
			loginURL := buildLoginURL(buildSignURL(h.baseURL, docID))
			http.Redirect(w, r, loginURL, http.StatusFound)
			return
		}
		HandleError(w, err)
		return
	}

	docID, err := validateDocID(r)
	if err != nil {
		HandleError(w, models.ErrInvalidDocument)
		return
	}

	ctx := r.Context()

	var referer *string
	if referrerParam := r.FormValue("referrer"); referrerParam != "" {
		referer = &referrerParam
	} else if referrerParam := r.URL.Query().Get("referrer"); referrerParam != "" {
		referer = &referrerParam
	} else {
		fmt.Printf("DEBUG: No referrer found in form or URL\n")
	}

	request := &models.SignatureRequest{
		DocID:   docID,
		User:    user,
		Referer: referer,
	}

	err = h.signatureService.CreateSignature(ctx, request)
	if err != nil {
		if errors.Is(err, models.ErrSignatureAlreadyExists) {
			http.Redirect(w, r, buildSignURL(h.baseURL, docID), http.StatusFound)
			return
		}
		HandleError(w, err)
		return
	}

	http.Redirect(w, r, buildSignURL(h.baseURL, docID), http.StatusFound)
}

// HandleStatusJSON returns signature status as JSON
func (h *SignatureHandlers) HandleStatusJSON(w http.ResponseWriter, r *http.Request) {
	docID, err := validateDocID(r)
	if err != nil {
		HandleError(w, models.ErrInvalidDocument)
		return
	}

	ctx := r.Context()
	signatures, err := h.signatureService.GetDocumentSignatures(ctx, docID)
	if err != nil {
		HandleError(w, err)
		return
	}

	// Convert to JSON response format
	response := make([]map[string]interface{}, 0, len(signatures))
	for _, sig := range signatures {
		sigData := map[string]interface{}{
			"id":         sig.ID,
			"doc_id":     sig.DocID,
			"user_sub":   sig.UserSub,
			"user_email": sig.UserEmail,
			"signed_at":  sig.SignedAtUTC,
		}

		if sig.UserName != nil && *sig.UserName != "" {
			sigData["user_name"] = *sig.UserName
		}

		if serviceInfo := sig.GetServiceInfo(); serviceInfo != nil {
			sigData["service"] = map[string]interface{}{
				"name": serviceInfo.Name,
				"icon": serviceInfo.Icon,
				"type": serviceInfo.Type,
			}
		}

		response = append(response, sigData)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleUserSignatures displays the user's signatures page
func (h *SignatureHandlers) HandleUserSignatures(w http.ResponseWriter, r *http.Request) {
	user, err := h.userService.GetUser(r)
	if err != nil {
		HandleError(w, err)
		return
	}

	ctx := r.Context()
	signatures, err := h.signatureService.GetUserSignatures(ctx, user)
	if err != nil {
		HandleError(w, err)
		return
	}

	h.render(w, r, "signatures", PageData{User: user, BaseURL: h.baseURL, Signatures: signatures})
}

// render executes template with data
func (h *SignatureHandlers) render(w http.ResponseWriter, _ *http.Request, templateName string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if data.Year == 0 {
		data.Year = time.Now().Year()
	}
	if data.TemplateName == "" {
		data.TemplateName = templateName
	}
	if !data.IsAdmin {
		data.IsAdmin = admin.IsAdminUser(data.User)
	}

	templateData := map[string]interface{}{
		"User":         data.User,
		"Year":         data.Year,
		"DocID":        data.DocID,
		"Already":      data.Already,
		"SignedAt":     data.SignedAt,
		"TemplateName": data.TemplateName,
		"BaseURL":      data.BaseURL,
		"Signatures":   data.Signatures,
		"ServiceInfo":  data.ServiceInfo,
		"IsAdmin":      data.IsAdmin,
	}

	if err := h.template.ExecuteTemplate(w, "base", templateData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
