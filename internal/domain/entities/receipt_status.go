package entities

type ReceiptStatus string

const (
	ReceiptStatusPending     ReceiptStatus = "PENDING"
	ReceiptStatusNeedsReview ReceiptStatus = "NEEDS_REVIEW"
	ReceiptStatusConfirmed   ReceiptStatus = "CONFIRMED"
	ReceiptStatusRejected    ReceiptStatus = "REJECTED"
)

func CanTransition(from, to ReceiptStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case ReceiptStatusPending:
		return to == ReceiptStatusNeedsReview || to == ReceiptStatusRejected
	case ReceiptStatusNeedsReview:
		return to == ReceiptStatusConfirmed || to == ReceiptStatusRejected
	default:
		return false
	}
}
