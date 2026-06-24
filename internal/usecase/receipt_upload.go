package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type ReceiptUploadUseCase struct {
	repo    ports.ReceiptRepository
	storage ports.StorageProvider
	queue   ports.OCRQueue
}

type UploadInput struct {
	UserID string
	Data   []byte
}

type UploadResult struct {
	ReceiptID string `json:"receipt_id"`
	Status    string `json:"status"`
	Duplicate bool   `json:"duplicate"`
}

func NewReceiptUploadUseCase(repo ports.ReceiptRepository, storage ports.StorageProvider, queue ports.OCRQueue) *ReceiptUploadUseCase {
	return &ReceiptUploadUseCase{repo: repo, storage: storage, queue: queue}
}

func (u *ReceiptUploadUseCase) Execute(ctx context.Context, in UploadInput) (UploadResult, error) {
	if in.UserID == "" {
		return UploadResult{}, fmt.Errorf("user id is required")
	}
	if len(in.Data) == 0 {
		return UploadResult{}, fmt.Errorf("receipt image is required")
	}

	hash := sha256.Sum256(in.Data)
	imageHash := hex.EncodeToString(hash[:])

	existing, err := u.repo.FindByUserAndImageHash(ctx, in.UserID, imageHash)
	if err != nil {
		return UploadResult{}, err
	}
	if existing != nil {
		return UploadResult{ReceiptID: existing.ID, Status: string(existing.Status), Duplicate: true}, nil
	}

	objectName, err := randomObjectName()
	if err != nil {
		return UploadResult{}, err
	}
	url, err := u.storage.Upload(ctx, objectName, in.Data)
	if err != nil {
		return UploadResult{}, err
	}

	receipt, err := u.repo.CreatePendingReceipt(ctx, in.UserID, url, imageHash)
	if err != nil {
		return UploadResult{}, err
	}

	if err := u.queue.Enqueue(ctx, receipt.ID); err != nil {
		return UploadResult{}, err
	}

	return UploadResult{ReceiptID: receipt.ID, Status: string(entities.ReceiptStatusPending), Duplicate: false}, nil
}

func randomObjectName() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "receipts/" + hex.EncodeToString(b) + ".jpg", nil
}
