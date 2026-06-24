package usecase

import (
	"context"
	"fmt"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type ReceiptProcessUseCase struct {
	repo ports.ReceiptRepository
	ocr  ports.OCRProvider
}

func NewReceiptProcessUseCase(repo ports.ReceiptRepository, ocr ports.OCRProvider) *ReceiptProcessUseCase {
	return &ReceiptProcessUseCase{repo: repo, ocr: ocr}
}

func (u *ReceiptProcessUseCase) Execute(ctx context.Context, receiptID string) error {
	receipt, err := u.repo.GetByID(ctx, receiptID)
	if err != nil {
		return err
	}
	if receipt == nil {
		return fmt.Errorf("receipt not found")
	}

	raw, err := u.ocr.Extract(ctx, receipt.ImageURL)
	if err != nil {
		return err
	}

	summary := ParseOCRText(raw)
	summary.ReceiptID = receipt.ID
	summary.Status = entities.ReceiptStatusNeedsReview
	if summary.Store.Name == "" {
		summary.Store.Name = "Unknown"
	}

	return u.repo.MarkNeedsReview(ctx, receipt.ID, summary)
}
