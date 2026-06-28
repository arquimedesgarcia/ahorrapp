package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"ahorrapp/internal/domain/ports"
)

const maxWorkerAttempts = 3

type retryCapableQueue interface {
	RequeueWithBackoff(ctx context.Context, receiptID string, attempt int, reason string) error
	SendToDeadLetter(ctx context.Context, receiptID, reason string) error
}

type ReceiptWorker struct {
	queue   ports.OCRQueue
	process *ReceiptProcessUseCase
	attempt map[string]int
}

func NewReceiptWorker(queue ports.OCRQueue, process *ReceiptProcessUseCase) *ReceiptWorker {
	return &ReceiptWorker{queue: queue, process: process, attempt: make(map[string]int)}
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
			log.Printf("event=ocr_worker_dequeue_failed error=%q", err)
			continue
		}

		log.Printf("event=ocr_worker_job_received receipt_id=%s", receiptID)
		if err := w.process.Execute(ctx, receiptID); err != nil {
			attempt := w.attempt[receiptID] + 1
			w.attempt[receiptID] = attempt
			reason := err.Error()
			log.Printf("event=ocr_worker_job_failed receipt_id=%s attempt=%d error=%q", receiptID, attempt, reason)

			retryQueue, ok := w.queue.(retryCapableQueue)
			if !ok {
				continue
			}

			if attempt >= maxWorkerAttempts {
				if dlqErr := retryQueue.SendToDeadLetter(ctx, receiptID, reason); dlqErr != nil {
					log.Printf("event=ocr_worker_dlq_failed receipt_id=%s error=%q", receiptID, dlqErr)
					continue
				}
				delete(w.attempt, receiptID)
				log.Printf("event=ocr_worker_sent_to_dlq receipt_id=%s attempts=%d", receiptID, attempt)
				continue
			}

			if requeueErr := retryQueue.RequeueWithBackoff(ctx, receiptID, attempt, reason); requeueErr != nil {
				log.Printf("event=ocr_worker_requeue_failed receipt_id=%s attempt=%d error=%q", receiptID, attempt, requeueErr)
				continue
			}
			log.Printf("event=ocr_worker_requeued receipt_id=%s attempt=%d backoff=%s", receiptID, attempt, fmt.Sprintf("%ds", attempt*attempt))
			continue
		}

		delete(w.attempt, receiptID)
		log.Printf("event=ocr_worker_job_processed receipt_id=%s status=success", receiptID)
	}
}
