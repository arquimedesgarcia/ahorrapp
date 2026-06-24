package usecase

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	"ahorrapp/internal/domain/entities"
)

var (
	reStoreLine = regexp.MustCompile(`(?i)^(super|market|store|tienda|supermarket).*`)
	reDate      = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	rePriceLine = regexp.MustCompile(`(?i)^(.*)\s+(\d+(?:\.\d+)?)\s*x\s*(\d+(?:\.\d+)?)\s+(USD|Bs\.?|VES)$`)
)

func ParseOCRText(raw string) entities.EditableSummary {
	summary := entities.EditableSummary{
		Status: entities.ReceiptStatusNeedsReview,
		Store:  entities.StoreSummary{},
		Items:  make([]entities.EditableItem, 0),
	}

	s := bufio.NewScanner(strings.NewReader(raw))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		if summary.Store.Name == "" && reStoreLine.MatchString(line) {
			summary.Store.Name = line
			continue
		}

		if summary.PurchaseDate == nil {
			if d := reDate.FindString(line); d != "" {
				summary.PurchaseDate = &d
				continue
			}
		}

		if m := rePriceLine.FindStringSubmatch(line); len(m) == 5 {
			qty := parseFloatPtr(m[2])
			price := parseFloatPtr(m[3])
			cur := m[4]
			summary.Items = append(summary.Items, entities.EditableItem{
				RawText:   strings.TrimSpace(m[1]),
				Quantity:  qty,
				UnitPrice: price,
				Currency:  &cur,
			})
		}
	}

	return summary
}

func parseFloatPtr(in string) *float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(in), 64)
	if err != nil {
		return nil
	}
	return &f
}
