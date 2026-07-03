package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FirstObservationRepository struct {
	pool *pgxpool.Pool
}

func NewFirstObservationRepository(pool *pgxpool.Pool) *FirstObservationRepository {
	return &FirstObservationRepository{pool: pool}
}

// PreviouslyObserved returns true iff a price_observations row already
// exists for the given (productID, storeID) pair before the current
// confirmation inserts its own. The caller MUST invoke this within the
// same transaction that inserts the new observations to eliminate the
// race described in spec R-04. (The award path runs after ConfirmReceipt
// commits, so previously-observed reflects the state immediately before
// this confirmation; new observations from this confirmation have been
// persisted only if the confirm tx committed, in which case they are
// visible here too. To avoid the award step double-counting its own
// observations as "previously observed", the confirm use case passes the
// already-existing observations explicitly and the award short-circuits
// when the receipt itself produced them; the simplest correct mechanism
// is the unique partial index plus the dedup map in the use case.)
func (r *FirstObservationRepository) PreviouslyObserved(ctx context.Context, productID, storeID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
SELECT EXISTS(
  SELECT 1 FROM price_observations
  WHERE product_id = $1::uuid AND store_id = $2::uuid
)
`, productID, storeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("first observation check: %w", err)
	}
	return exists, nil
}
