package usecase

// Reason codes written verbatim into loyalty_transactions.reason.
// When a single award row accumulates several reasons they are joined
// with ";". See specs/006-loyalty-points/data-model.md.
const (
	ReasonReceiptConfirmed        = "receipt_confirmed"
	ReasonFirstObservationProduct = "first_observation_product"
	ReasonFirstObservationStore   = "first_observation_store"
	ReasonDataCompletion          = "data_completion"
	ReasonDailyLimitReached       = "daily_limit_reached"
)

const loyaltyHistoryLimit = 100
