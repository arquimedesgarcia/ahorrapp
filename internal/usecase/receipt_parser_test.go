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

func TestParseOCRText_VenezuelanSplitLines(t *testing.T) {
	raw := "FARMATODO\nDIRECCION: AV PRINCIPAL\nFACTURA\nFECHA:02-07-2026\n" +
		"1xCLORO\n" +
		"Bs 415,81\n" +
		"2 x JABON LAVAPLATOS\n" +
		"Bs 703,67\n" +
		"1xSUAVIZANTE SUAVITEL\n" +
		"Bs 537,35\n" +
		"TOTAL\n" +
		"Bs 1.656,83\n"

	out := ParseOCRText(raw)

	if out.PurchaseDate == nil || *out.PurchaseDate != "2026-07-02" {
		t.Fatalf("expected 2026-07-02, got %+v", out.PurchaseDate)
	}
	if len(out.Items) != 3 {
		t.Fatalf("expected 3 items, got %d (%v)", len(out.Items), out.Items)
	}
	if out.Items[0].RawText != "CLORO" {
		t.Errorf("expected CLORO, got %q", out.Items[0].RawText)
	}
	if out.Items[0].UnitPrice == nil || *out.Items[0].UnitPrice != 415.81 {
		t.Errorf("expected 415.81, got %+v", out.Items[0].UnitPrice)
	}
	if out.Items[0].Currency == nil || *out.Items[0].Currency != "Bs" {
		t.Errorf("expected Bs, got %+v", out.Items[0].Currency)
	}
	if out.Total == nil || *out.Total != 1656.83 {
		t.Errorf("expected total 1656.83, got %+v", out.Total)
	}
}

func TestParseVenezuelanPrice(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"1.656,83", 1656.83},
		{"415,81", 415.81},
		{"123", 123},
		{"1.656.83", 1656.83},
		{"0,50", 0.50},
	}
	for _, c := range cases {
		got, err := parseVenezuelanPrice(c.in)
		if err != nil {
			t.Errorf("parseVenezuelanPrice(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseVenezuelanPrice(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseOCRText_PriceBeforeItem(t *testing.T) {
	raw := "FACTURA\nFECHA:02-07-2026\n" +
		"Bs415,81\n" +
		"1xCLORO\n" +
		"1xJABON\n" +
		"Bs 703,67\n"

	out := ParseOCRText(raw)
	if len(out.Items) != 2 {
		t.Fatalf("expected 2 items, got %d (%+v)", len(out.Items), out.Items)
	}
	if out.Items[0].RawText != "CLORO" {
		t.Errorf("item[0] expected CLORO, got %q", out.Items[0].RawText)
	}
	if out.Items[0].UnitPrice == nil || *out.Items[0].UnitPrice != 415.81 {
		t.Errorf("item[0] price expected 415.81, got %+v", out.Items[0].UnitPrice)
	}
	if out.Items[1].RawText != "JABON" {
		t.Errorf("item[1] expected JABON, got %q", out.Items[1].RawText)
	}
	if out.Items[1].UnitPrice == nil || *out.Items[1].UnitPrice != 703.67 {
		t.Errorf("item[1] price expected 703.67, got %+v", out.Items[1].UnitPrice)
	}
}
