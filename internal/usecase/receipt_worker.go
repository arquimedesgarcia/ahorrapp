package usecase

import (
	"context"
	"errors"
	"log"

	"ahorrapp/internal/domain/ports"
)

type ReceiptWorker struct {
	queue   ports.OCRQueue
	process *ReceiptProcessUseCase
}

func NewReceiptWorker(queue ports.OCRQueue, process *ReceiptProcessUseCase) *ReceiptWorker {
	return &ReceiptWorker{queue: queue, process: process}
}

func (w *ReceiptWorker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		receiptID, err := w.queue.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			log.Printf("worker dequeue failed: %v", err)
			continue
		}

		if err := w.process.Execute(ctx, receiptID); err != nil {
			log.Printf("worker process receipt %s failed: %v", receiptID, err)
		}
	}
}
