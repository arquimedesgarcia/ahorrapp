package ports

import "context"

// FirstObservationChecker reports whether a (product, store) pair already
// has at least one price_observations row before the current confirmation
// inserts its own. Used to decide on the "first_observation_product"
// bonus. The check MUST run inside the same transaction that inserts the
// new observations so the race described in spec R-04 is eliminated.
type FirstObservationChecker interface {
	PreviouslyObserved(ctx context.Context, productID, storeID string) (bool, error)
}
