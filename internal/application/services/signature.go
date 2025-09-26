package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/pkg/crypto"
	"github.com/btouchard/ackify-ce/pkg/logger"
)

type repository interface {
    Create(ctx context.Context, signature *models.Signature) error
    GetByDocAndUser(ctx context.Context, docID, userSub string) (*models.Signature, error)
    GetByDoc(ctx context.Context, docID string) ([]*models.Signature, error)
    GetByUser(ctx context.Context, userSub string) ([]*models.Signature, error)
    ExistsByDocAndUser(ctx context.Context, docID, userSub string) (bool, error)
    CheckUserSignatureStatus(ctx context.Context, docID, userIdentifier string) (bool, error)
    GetLastSignature(ctx context.Context) (*models.Signature, error)
    GetAllSignaturesOrdered(ctx context.Context) ([]*models.Signature, error)
    UpdatePrevHash(ctx context.Context, id int64, prevHash *string) error
}

type cryptoSigner interface {
	CreateSignature(docID string, user *models.User, timestamp time.Time, nonce string) (string, string, error)
}

type SignatureService struct {
	repo   repository
	signer cryptoSigner
}

// NewSignatureService creates a new signature service
func NewSignatureService(repo repository, signer cryptoSigner) *SignatureService {
	return &SignatureService{
		repo:   repo,
		signer: signer,
	}
}

func (s *SignatureService) CreateSignature(ctx context.Context, request *models.SignatureRequest) error {
	if request.User == nil || !request.User.IsValid() {
		return models.ErrInvalidUser
	}

	if request.DocID == "" {
		return models.ErrInvalidDocument
	}

	exists, err := s.repo.ExistsByDocAndUser(ctx, request.DocID, request.User.Sub)
	if err != nil {
		return fmt.Errorf("failed to check existing signature: %w", err)
	}

	if exists {
		return models.ErrSignatureAlreadyExists
	}

	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	timestamp := time.Now().UTC()
	payloadHash, signatureB64, err := s.signer.CreateSignature(request.DocID, request.User, timestamp, nonce)
	if err != nil {
		return fmt.Errorf("failed to create cryptographic signature: %w", err)
	}

	lastSignature, err := s.repo.GetLastSignature(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last signature for chaining: %w", err)
	}

	var prevHashB64 *string
	if lastSignature != nil {
		hash := lastSignature.ComputeRecordHash()
		prevHashB64 = &hash
		logger.Logger.Info("Chaining to previous signature",
			"prevID", lastSignature.ID,
			"prevHash", hash[:16]+"...")
	} else {
		logger.Logger.Info("Creating genesis signature (no previous signature)")
	}

	var userName *string
	if request.User.Name != "" {
		userName = &request.User.Name
	}

	logger.Logger.Info("Creating signature",
		"docID", request.DocID,
		"userSub", request.User.Sub,
		"userEmail", request.User.NormalizedEmail(),
		"userName", request.User.Name)

	signature := &models.Signature{
		DocID:       request.DocID,
		UserSub:     request.User.Sub,
		UserEmail:   request.User.NormalizedEmail(),
		UserName:    userName,
		SignedAtUTC: timestamp,
		PayloadHash: payloadHash,
		Signature:   signatureB64,
		Nonce:       nonce,
		Referer:     request.Referer,
		PrevHash:    prevHashB64,
	}

	if err := s.repo.Create(ctx, signature); err != nil {
		return fmt.Errorf("failed to save signature: %w", err)
	}

	logger.Logger.Info("Signature created successfully", "id", signature.ID)

	return nil
}

func (s *SignatureService) GetSignatureStatus(ctx context.Context, docID string, user *models.User) (*models.SignatureStatus, error) {
	if user == nil || !user.IsValid() {
		return nil, models.ErrInvalidUser
	}

	signature, err := s.repo.GetByDocAndUser(ctx, docID, user.Sub)
	if err != nil {
		if errors.Is(err, models.ErrSignatureNotFound) {
			return &models.SignatureStatus{
				DocID:     docID,
				UserEmail: user.Email,
				IsSigned:  false,
				SignedAt:  nil,
			}, nil
		}
		return nil, fmt.Errorf("failed to get signature: %w", err)
	}

	return &models.SignatureStatus{
		DocID:     docID,
		UserEmail: user.Email,
		IsSigned:  true,
		SignedAt:  &signature.SignedAtUTC,
	}, nil
}

func (s *SignatureService) GetDocumentSignatures(ctx context.Context, docID string) ([]*models.Signature, error) {
	signatures, err := s.repo.GetByDoc(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document signatures: %w", err)
	}

	return signatures, nil
}

func (s *SignatureService) GetUserSignatures(ctx context.Context, user *models.User) ([]*models.Signature, error) {
	if user == nil || !user.IsValid() {
		return nil, models.ErrInvalidUser
	}

	signatures, err := s.repo.GetByUser(ctx, user.Sub)
	if err != nil {
		return nil, fmt.Errorf("failed to get user signatures: %w", err)
	}

	return signatures, nil
}

func (s *SignatureService) GetSignatureByDocAndUser(ctx context.Context, docID string, user *models.User) (*models.Signature, error) {
	if user == nil || !user.IsValid() {
		return nil, models.ErrInvalidUser
	}

	signature, err := s.repo.GetByDocAndUser(ctx, docID, user.Sub)
	if err != nil {
		return nil, fmt.Errorf("failed to get signature: %w", err)
	}

	return signature, nil
}

func (s *SignatureService) CheckUserSignature(ctx context.Context, docID, userIdentifier string) (bool, error) {
	exists, err := s.repo.CheckUserSignatureStatus(ctx, docID, userIdentifier)
	if err != nil {
		return false, fmt.Errorf("failed to check user signature: %w", err)
	}

	return exists, nil
}

type ChainIntegrityResult struct {
	IsValid      bool
	TotalRecords int
	BreakAtID    *int64
	Details      string
}

func (s *SignatureService) VerifyChainIntegrity(ctx context.Context) (*ChainIntegrityResult, error) {
	signatures, err := s.repo.GetAllSignaturesOrdered(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures for chain verification: %w", err)
	}

	result := &ChainIntegrityResult{
		IsValid:      true,
		TotalRecords: len(signatures),
	}

	if len(signatures) == 0 {
		result.Details = "No signatures found"
		return result, nil
	}

	if signatures[0].PrevHash != nil {
		result.IsValid = false
		result.BreakAtID = &signatures[0].ID
		result.Details = "Genesis signature has non-null previous hash"
		return result, nil
	}

	for i := 1; i < len(signatures); i++ {
		current := signatures[i]
		previous := signatures[i-1]

		expectedHash := previous.ComputeRecordHash()

		if current.PrevHash == nil {
			result.IsValid = false
			result.BreakAtID = &current.ID
			result.Details = fmt.Sprintf("Signature %d has null previous hash, expected: %s...", current.ID, expectedHash[:16])
			return result, nil
		}

		if *current.PrevHash != expectedHash {
			result.IsValid = false
			result.BreakAtID = &current.ID
			result.Details = fmt.Sprintf("Hash mismatch at signature %d: expected %s..., got %s...",
				current.ID, expectedHash[:16], (*current.PrevHash)[:16])
			return result, nil
		}
	}

	result.Details = "Chain integrity verified successfully"
	return result, nil
}

// RebuildChain reconstructs the hash chain for existing signatures
// This should be used once after deploying the chain feature to populate prev_hash
func (s *SignatureService) RebuildChain(ctx context.Context) error {
	signatures, err := s.repo.GetAllSignaturesOrdered(ctx)
	if err != nil {
		return fmt.Errorf("failed to get signatures for chain rebuild: %w", err)
	}

	if len(signatures) == 0 {
		logger.Logger.Info("No signatures found, nothing to rebuild")
		return nil
	}

	logger.Logger.Info("Starting chain rebuild", "totalSignatures", len(signatures))

    if signatures[0].PrevHash != nil {
        if err := s.repo.UpdatePrevHash(ctx, signatures[0].ID, nil); err != nil {
            logger.Logger.Warn("Failed to nullify genesis prev_hash", "id", signatures[0].ID, "error", err)
        }
    }

	for i := 1; i < len(signatures); i++ {
		current := signatures[i]
		previous := signatures[i-1]

		expectedHash := previous.ComputeRecordHash()

        if current.PrevHash == nil || *current.PrevHash != expectedHash {
            logger.Logger.Info("Chain rebuild: updating prev_hash",
                "id", current.ID,
                "expectedHash", expectedHash[:16]+"...",
                "hadPrevHash", current.PrevHash != nil)
            if err := s.repo.UpdatePrevHash(ctx, current.ID, &expectedHash); err != nil {
                logger.Logger.Warn("Failed to update prev_hash", "id", current.ID, "error", err)
            }
        }
    }

	logger.Logger.Info("Chain rebuild completed", "processedSignatures", len(signatures))
	return nil
}
