package usecase

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"ahorrapp/internal/domain/entities"
)

var (
	reItemLine   = regexp.MustCompile(`(?i)^\s*(\d+)\s*x\s*(.+?)\s*$`)
	rePriceOnly  = regexp.MustCompile(`(?i)^\s*Bs\.?\s*([\d.,]+)\s*$`)
	rePriceLine  = regexp.MustCompile(`(?i)^(.*?)\s+(\d+(?:\.\d+)?)\s*x\s*(\d+(?:[.,]\d+)?)\s+(USD|Bs\.?|VES)$`)
	reDateInLine = regexp.MustCompile(`\b(\d{2}[-/]\d{2}[-/]\d{4}|\d{4}[-/]\d{2}[-/]\d{2})\b`)
	reFACTURA    = regexp.MustCompile(`(?i)^\s*FACTURA\s*:?\s*$`)
	reSkipLine   = regexp.MustCompile(`(?i)^\s*(DIRECCION|DIRECCIÓN|USUARIO|CAJA|HORA|NIT|RIF|TELF|TELÉFONO|FECHA\s*:?)\s*[:.]?`)
)

var dateLayouts = []string{
	"02-01-2006",
	"02/01/2006",
	"2006-01-02",
	"2006/01/02",
}

func parseVenezuelanPrice(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
		return strconv.ParseFloat(s, 64)
	}
	parts := strings.Split(s, ".")
	if len(parts) >= 2 && len(parts[len(parts)-1]) <= 2 {
		last := parts[len(parts)-1]
		rest := strings.Join(parts[:len(parts)-1], "")
		return strconv.ParseFloat(rest+"."+last, 64)
	}
	return strconv.ParseFloat(s, 64)
}

func extractPrice(s string) *float64 {
	m := rePriceOnly.FindStringSubmatch(s)
	if len(m) != 2 {
		return nil
	}
	f, err := parseVenezuelanPrice(m[1])
	if err != nil {
		return nil
	}
	return &f
}

func extractDate(s string) *string {
	m := reDateInLine.FindStringSubmatch(s)
	if len(m) != 2 {
		return nil
	}
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, m[1]); err == nil {
			normalized := t.Format("2006-01-02")
			return &normalized
		}
	}
	return nil
}

func extractStore(lines []string) string {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if reFACTURA.MatchString(line) {
			return ""
		}
		if reSkipLine.MatchString(line) {
			continue
		}
		letterCount := 0
		hasSymbol := false
		for _, c := range line {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				letterCount++
			}
			if c == '?' || c == ':' || c == ';' {
				hasSymbol = true
			}
		}
		if letterCount >= 4 && !hasSymbol {
			return line
		}
	}
	return ""
}

func ParseOCRText(raw string) entities.EditableSummary {
	summary := entities.EditableSummary{
		Status: entities.ReceiptStatusNeedsReview,
		Store:  entities.StoreSummary{},
		Items:  make([]entities.EditableItem, 0),
	}

	lines := strings.Split(raw, "\n")
	defaultCur := "Bs"

	summary.Store.Name = extractStore(lines)

	var pending *entities.EditableItem
	var pendingPrice *float64
	flush := func() {
		if pending != nil {
			summary.Items = append(summary.Items, *pending)
			pending = nil
		}
	}

	for _, lineRaw := range lines {
		line := strings.TrimSpace(lineRaw)
		if line == "" {
			continue
		}

		if summary.PurchaseDate == nil {
			if d := extractDate(line); d != nil {
				summary.PurchaseDate = d
				continue
			}
		}

		if m := rePriceLine.FindStringSubmatch(line); len(m) == 5 {
			qty, _ := strconv.ParseFloat(m[2], 64)
			price, _ := parseVenezuelanPrice(m[3])
			cur := m[4]
			summary.Items = append(summary.Items, entities.EditableItem{
				RawText:   strings.TrimSpace(m[1]),
				Quantity:  &qty,
				UnitPrice: &price,
				Currency:  &cur,
			})
			continue
		}

		if m := reItemLine.FindStringSubmatch(line); len(m) == 3 {
			qty, _ := strconv.ParseFloat(m[1], 64)
			desc := strings.TrimSpace(m[2])
			flush()
			pending = &entities.EditableItem{
				RawText:  desc,
				Quantity: &qty,
				Currency: &defaultCur,
			}
			if pendingPrice != nil {
				pending.UnitPrice = pendingPrice
				pendingPrice = nil
			}
			continue
		}

		if price := extractPrice(line); price != nil {
			if pending != nil {
				pending.UnitPrice = price
				flush()
			} else {
				pendingPrice = price
			}
			continue
		}

		if summary.Total == nil && pendingPrice != nil {
			summary.Total = pendingPrice
			pendingPrice = nil
		}
		flush()
	}

	if summary.Total == nil && pendingPrice != nil {
		summary.Total = pendingPrice
	}
	flush()

	return summary
}
