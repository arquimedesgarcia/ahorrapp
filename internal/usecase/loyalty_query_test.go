package usecase

import (
	"context"
	"testing"
	"time"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type fakeLoyaltyQueryRepo struct {
	balance   int
	balanceEr error
	history   []entities.LoyaltyTransaction
	historyEr error
	contrib   ports.ContributorStats
}

func (f *fakeLoyaltyQueryRepo) AwardForReceipt(context.Context, string, string, int, string) error {
	return nil
}
func (f *fakeLoyaltyQueryRepo) DailyGrantCount(context.Context, string) (int, error) {
	return 0, nil
}
func (f *fakeLoyaltyQueryRepo) Balance(context.Context, string) (int, error) {
	return f.balance, f.balanceEr
}
func (f *fakeLoyaltyQueryRepo) History(context.Context, string, int) ([]entities.LoyaltyTransaction, error) {
	return f.history, f.historyEr
}
func (f *fakeLoyaltyQueryRepo) ContributorStats(context.Context, string) (ports.ContributorStats, error) {
	return f.contrib, nil
}

var _ ports.LoyaltyRepository = (*fakeLoyaltyQueryRepo)(nil)

func TestLoyaltyQuery_ConsistentBalance(t *testing.T) {
	r := &fakeLoyaltyQueryRepo{
		balance: 18,
		history: []entities.LoyaltyTransaction{
			{ID: "t1", UserID: "u-1", Points: 18, Reason: "receipt_confirmed;data_completion", CreatedAt: time.Now()},
		},
	}
	uc := NewLoyaltyQueryUseCase(r)
	resp, err := uc.GetLoyalty(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("GetLoyalty: %v", err)
	}
	if resp.Balance != 18 {
		t.Errorf("Balance = %d, want 18", resp.Balance)
	}
	if len(resp.History) != 1 {
		t.Fatalf("expected 1 movement, got %d", len(resp.History))
	}
	if resp.History[0].Points != 18 {
		t.Errorf("history[0].Points = %d, want 18", resp.History[0].Points)
	}
	if resp.History[0].Reason != "receipt_confirmed;data_completion" {
		t.Errorf("history[0].Reason = %q", resp.History[0].Reason)
	}
}

func TestLoyaltyQuery_EmptyHistory(t *testing.T) {
	r := &fakeLoyaltyQueryRepo{balance: 0, history: nil}
	uc := NewLoyaltyQueryUseCase(r)
	resp, err := uc.GetLoyalty(context.Background(), "fresh-user")
	if err != nil {
		t.Fatalf("GetLoyalty: %v", err)
	}
	if resp.Balance != 0 {
		t.Errorf("Balance = %d, want 0", resp.Balance)
	}
	if resp.History == nil {
		t.Errorf("History must be non-nil empty slice, got nil")
	}
	if len(resp.History) != 0 {
		t.Errorf("len(History) = %d, want 0", len(resp.History))
	}
}

func TestLoyaltyQuery_OrdersByCreatedAtDesc(t *testing.T) {
	now := time.Now()
	later := now.Add(1 * time.Hour)
	// The repository contract is to return movements in created_at DESC
	// order. The fake mocks that contract by emitting `later` first.
	r := &fakeLoyaltyQueryRepo{
		balance: 5,
		history: []entities.LoyaltyTransaction{
			{ID: "newer", CreatedAt: later},
			{ID: "older", CreatedAt: now},
		},
	}
	uc := NewLoyaltyQueryUseCase(r)
	resp, err := uc.GetLoyalty(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("GetLoyalty: %v", err)
	}
	// The query repo returns them in created_at DESC order; the use case
	// must preserve that order.
	if resp.History[0].ID != "newer" {
		t.Errorf("expected newer first, got %q", resp.History[0].ID)
	}
}

func TestLoyaltyQuery_ReceiptIDPresentWhenSet(t *testing.T) {
	rid := "r-abc"
	r := &fakeLoyaltyQueryRepo{
		balance: 0,
		history: []entities.LoyaltyTransaction{
			{ID: "t1", ReceiptID: &rid},
		},
	}
	uc := NewLoyaltyQueryUseCase(r)
	resp, err := uc.GetLoyalty(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("GetLoyalty: %v", err)
	}
	if resp.History[0].ReceiptID == nil || *resp.History[0].ReceiptID != rid {
		t.Errorf("ReceiptID = %v, want %q", resp.History[0].ReceiptID, rid)
	}
}

func TestLoyaltyQuery_LevelThresholds(t *testing.T) {
	cases := []struct {
		points int
		want   string
	}{
		{0, "Bronce"},
		{18, "Bronce"},
		{49, "Bronce"},
		{50, "Plata"},
		{199, "Plata"},
		{200, "Oro"},
		{499, "Oro"},
		{500, "Platino"},
		{9999, "Platino"},
	}
	for _, c := range cases {
		got := levelFor(c.points)
		if got != c.want {
			t.Errorf("levelFor(%d) = %q, want %q", c.points, got, c.want)
		}
	}
}

func TestLoyaltyQuery_IncludesContributorStats(t *testing.T) {
	r := &fakeLoyaltyQueryRepo{
		balance: 75,
		contrib: ports.ContributorStats{
			ReceiptsConfirmed: 5,
			PriceObservations: 14,
			UniqueStores:      3,
			UniqueProducts:    9,
		},
	}
	uc := NewLoyaltyQueryUseCase(r)
	resp, err := uc.GetLoyalty(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("GetLoyalty: %v", err)
	}
	if resp.Level != "Plata" {
		t.Errorf("Level = %q, want Plata (75 pts)", resp.Level)
	}
	if resp.Contributor.ReceiptsConfirmed != 5 {
		t.Errorf("ReceiptsConfirmed = %d, want 5", resp.Contributor.ReceiptsConfirmed)
	}
	if resp.Contributor.UniqueProducts != 9 {
		t.Errorf("UniqueProducts = %d, want 9", resp.Contributor.UniqueProducts)
	}
}
