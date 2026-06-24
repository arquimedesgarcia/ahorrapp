package events

import (
	"context"
	"log"
)

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) EmitReceiptConfirmed(ctx context.Context, receiptID, userID string, observations int) error {
	_ = ctx
	log.Printf("event=receipt_confirmed receipt_id=%s user_id=%s observations=%d", receiptID, userID, observations)
	return nil
}
