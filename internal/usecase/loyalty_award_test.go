package usecase

import (
	"context"
	"strings"
	"testing"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

// --- fake LoyaltyRepository ---

type fakeLoyaltyRepo struct {
	awardedCalls  []awardCall
	dailyCount    int
	balance       int
	history       []entities.LoyaltyTransaction
	awardErr      error
	dailyCountErr error
	balanceErr    error
	historyErr    error
}

type awardCall struct {
	userID, receiptID string
	points            int
	reason            string
}

func (f *fakeLoyaltyRepo) AwardForReceipt(_ context.Context, userID, receiptID string, points int, reason string) error {
	f.awardedCalls = append(f.awardedCalls, awardCall{userID, receiptID, points, reason})
	if f.awardErr != nil {
		return f.awardErr
	}
	return nil
}
func (f *fakeLoyaltyRepo) DailyGrantCount(context.Context, string) (int, error) {
	return f.dailyCount, f.dailyCountErr
}
func (f *fakeLoyaltyRepo) Balance(context.Context, string) (int, error) {
	return f.balance, f.balanceErr
}
func (f *fakeLoyaltyRepo) History(context.Context, string, int) ([]entities.LoyaltyTransaction, error) {
	return f.history, f.historyErr
}
func (f *fakeLoyaltyRepo) ContributorStats(context.Context, string) (ports.ContributorStats, error) {
	return ports.ContributorStats{}, nil
}

// --- helpers ---

func awardUseCaseBase(repo ports.LoyaltyRepository) *LoyaltyAwardUseCase {
	return NewLoyaltyAwardUseCase(repo, 10, 5, 3, 20)
}

func sampleConfirmedReceipt() entities.Receipt {
	return entities.Receipt{
		ID:     "r-1",
		UserID: "u-1",
		Status: entities.ReceiptStatusConfirmed,
	}
}

func basePayload() entities.ConfirmPayload {
	cur := "Bs"
	qty := 1.0
	price := 1.0
	return entities.ConfirmPayload{
		Store:        entities.StoreSummary{Name: "Store"},
		PurchaseDate: "2026-06-24",
		// Total = 0 so dataCompleted() returns false and the +3 data
		// completion bonus does NOT stack on the base 10 in US1 tests.
		Total: 0,
		Items: []entities.EditableItem{{
			RawText:   "LECHE",
			Quantity:  &qty,
			UnitPrice: &price,
			Currency:  &cur,
		}},
	}
}

func fullPayload() entities.ConfirmPayload {
	p := basePayload()
	p.PurchaseDate = "2026-06-24"
	p.Total = 1.66
	return p
}

func baseInput(r entities.Receipt, p entities.ConfirmPayload) AwardInput {
	return AwardInput{Receipt: r, Payload: p}
}

// --- US1 tests ---

func TestLoyaltyAward_BaseOnceAwarded(t *testing.T) {
	repo := &fakeLoyaltyRepo{}
	uc := awardUseCaseBase(repo)

	_, _ = uc.AwardForReceipt(context.Background(), baseInput(sampleConfirmedReceipt(), basePayload()))
	if len(repo.awardedCalls) != 1 {
		t.Fatalf("expected 1 award call, got %d", len(repo.awardedCalls))
	}
	call := repo.awardedCalls[0]
	if call.points != 10 {
		t.Errorf("points = %d, want 10 (base only)", call.points)
	}
	if !strings.Contains(call.reason, ReasonReceiptConfirmed) {
		t.Errorf("reason %q must contain %q", call.reason, ReasonReceiptConfirmed)
	}
	if strings.Contains(call.reason, ReasonFirstObservationProduct) {
		t.Errorf("reason should not contain first_observation_product when count == 0")
	}
}

func TestLoyaltyAward_AlreadyAwardedSwallowed(t *testing.T) {
	repo := &fakeLoyaltyRepo{awardErr: ports.ErrAlreadyAwarded}
	uc := awardUseCaseBase(repo)

	_, _ = uc.AwardForReceipt(context.Background(), baseInput(sampleConfirmedReceipt(), basePayload()))
	if len(repo.awardedCalls) != 1 {
		t.Fatalf("expected exactly 1 call (the attempted insert), got %d", len(repo.awardedCalls))
	}
}

func TestLoyaltyAward_DailyLimit(t *testing.T) {
	repo := &fakeLoyaltyRepo{dailyCount: 20}
	uc := awardUseCaseBase(repo)

	_, _ = uc.AwardForReceipt(context.Background(), baseInput(sampleConfirmedReceipt(), basePayload()))
	if len(repo.awardedCalls) != 1 {
		t.Fatalf("expected 1 award row even when capped, got %d", len(repo.awardedCalls))
	}
	call := repo.awardedCalls[0]
	if call.points != 0 {
		t.Errorf("points = %d, want 0 when daily cap reached", call.points)
	}
	if call.reason != ReasonDailyLimitReached {
		t.Errorf("reason = %q, want %q", call.reason, ReasonDailyLimitReached)
	}
}

func TestLoyaltyAward_NonConfirmedReceiptIgnored(t *testing.T) {
	repo := &fakeLoyaltyRepo{}
	uc := awardUseCaseBase(repo)

	r := sampleConfirmedReceipt()
	r.Status = entities.ReceiptStatusPending
	_, _ = uc.AwardForReceipt(context.Background(), baseInput(r, basePayload()))
	if len(repo.awardedCalls) != 0 {
		t.Fatalf("expected 0 award calls for non-confirmed receipt, got %d", len(repo.awardedCalls))
	}
}

// --- US2 tests ---

func TestLoyaltyAward_FirstObservationProduct(t *testing.T) {
	repo := &fakeLoyaltyRepo{}
	uc := awardUseCaseBase(repo)

	in := baseInput(sampleConfirmedReceipt(), basePayload())
	in.FirstObservationPairCount = 2
	_, _ = uc.AwardForReceipt(context.Background(), in)
	call := repo.awardedCalls[0]
	if call.points != 10+5 {
		t.Errorf("points = %d, want %d (base + first-obs bonus once regardless of distinct pair count)", call.points, 10+5)
	}
	if !strings.Contains(call.reason, ReasonFirstObservationProduct) {
		t.Errorf("reason %q must contain %q", call.reason, ReasonFirstObservationProduct)
	}
}

func TestLoyaltyAward_FirstObservationStore(t *testing.T) {
	repo := &fakeLoyaltyRepo{}
	uc := awardUseCaseBase(repo)

	in := baseInput(sampleConfirmedReceipt(), basePayload())
	in.StoreCreatedNow = true
	_, _ = uc.AwardForReceipt(context.Background(), in)
	call := repo.awardedCalls[0]
	if call.points != 10+5 {
		t.Errorf("points = %d, want %d (base + store first-obs)", call.points, 10+5)
	}
	if !strings.Contains(call.reason, ReasonFirstObservationStore) {
		t.Errorf("reason %q must contain %q", call.reason, ReasonFirstObservationStore)
	}
	if strings.Contains(call.reason, ReasonFirstObservationProduct) {
		t.Errorf("reason should not contain first_observation_product when count == 0")
	}
}

func TestLoyaltyAward_DataCompletion(t *testing.T) {
	repo := &fakeLoyaltyRepo{}
	uc := awardUseCaseBase(repo)

	in := baseInput(sampleConfirmedReceipt(), fullPayload())
	_, _ = uc.AwardForReceipt(context.Background(), in)
	call := repo.awardedCalls[0]
	if !strings.Contains(call.reason, ReasonDataCompletion) {
		t.Errorf("reason %q must contain %q", call.reason, ReasonDataCompletion)
	}
	if call.points != 10+3 {
		t.Errorf("points = %d, want %d (base + data completion)", call.points, 10+3)
	}
}

func TestLoyaltyAward_DataCompletion_NotAwardedWhenMissing(t *testing.T) {
	cases := []struct {
		name    string
		payload entities.ConfirmPayload
	}{
		{"missing purchase_date", func() entities.ConfirmPayload { p := fullPayload(); p.PurchaseDate = ""; return p }()},
		{"zero total", func() entities.ConfirmPayload { p := fullPayload(); p.Total = 0; return p }()},
		{"item missing quantity", func() entities.ConfirmPayload { p := fullPayload(); p.Items[0].Quantity = nil; return p }()},
		{"item missing unit_price", func() entities.ConfirmPayload { p := fullPayload(); p.Items[0].UnitPrice = nil; return p }()},
		{"item missing currency", func() entities.ConfirmPayload { p := fullPayload(); p.Items[0].Currency = nil; return p }()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := &fakeLoyaltyRepo{}
			uc := awardUseCaseBase(repo)
			_, _ = uc.AwardForReceipt(context.Background(), baseInput(sampleConfirmedReceipt(), c.payload))
			call := repo.awardedCalls[0]
			if strings.Contains(call.reason, ReasonDataCompletion) {
				t.Errorf("case %q: reason should NOT contain data_completion, got %q", c.name, call.reason)
			}
			if call.points != 10 {
				t.Errorf("case %q: points = %d, want 10 (base only)", c.name, call.points)
			}
		})
	}
}

func TestLoyaltyAward_BonusesRespectDailyLimit(t *testing.T) {
	repo := &fakeLoyaltyRepo{dailyCount: 20}
	uc := awardUseCaseBase(repo)

	in := baseInput(sampleConfirmedReceipt(), fullPayload())
	in.FirstObservationPairCount = 5
	in.StoreCreatedNow = true
	_, _ = uc.AwardForReceipt(context.Background(), in)
	call := repo.awardedCalls[0]
	if call.points != 0 {
		t.Errorf("points = %d, want 0 when capped even if all bonuses apply", call.points)
	}
	if call.reason != ReasonDailyLimitReached {
		t.Errorf("reason = %q, want %q only (no bonus tokens when capped)", call.reason, ReasonDailyLimitReached)
	}
}

func TestLoyaltyAward_AllBonusesStacked(t *testing.T) {
	repo := &fakeLoyaltyRepo{}
	uc := awardUseCaseBase(repo)

	in := baseInput(sampleConfirmedReceipt(), fullPayload())
	in.FirstObservationPairCount = 1
	in.StoreCreatedNow = true
	_, _ = uc.AwardForReceipt(context.Background(), in)
	call := repo.awardedCalls[0]
	if call.points != 10+5+5+3 {
		t.Errorf("points = %d, want %d (base + first-prod + first-store + data-completion)", call.points, 10+5+5+3)
	}
	if !strings.Contains(call.reason, ReasonReceiptConfirmed) ||
		!strings.Contains(call.reason, ReasonFirstObservationProduct) ||
		!strings.Contains(call.reason, ReasonFirstObservationStore) ||
		!strings.Contains(call.reason, ReasonDataCompletion) {
		t.Errorf("reason %q must contain all four tokens", call.reason)
	}
}
