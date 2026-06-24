package usecase

import "testing"

func TestParseOCRText_ExtractsStoreDateAndItems(t *testing.T) {
	raw := "SUPERMARKET CENTRAL\nDATE 2026-06-21\nARROZ 1 x 2.40 USD\n"

	out := ParseOCRText(raw)

	if out.Store.Name != "SUPERMARKET CENTRAL" {
		t.Fatalf("expected store name, got %q", out.Store.Name)
	}
	if out.PurchaseDate == nil || *out.PurchaseDate != "2026-06-21" {
		t.Fatalf("expected parsed date, got %+v", out.PurchaseDate)
	}
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out.Items))
	}
	if out.Items[0].Currency == nil || *out.Items[0].Currency != "USD" {
		t.Fatalf("expected USD currency")
	}
}
