package entities

import "time"

type Receipt struct {
	ID           string
	UserID       string
	StoreID      *string
	ImageURL     string
	ImageHash    string
	Status       ReceiptStatus
	PurchaseDate *time.Time
	Total        *float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ReceiptListItem is the trimmed projection returned by the list endpoint.
// It deliberately omits the image_hash and the full image_url (which can
// be a long presigned S3 URL) to keep the list payload small for the
// mobile. The full EditableSummary is fetched lazily when the user
// taps an item.
type ReceiptListItem struct {
	ID           string        `json:"id"`
	Status       ReceiptStatus `json:"status"`
	StoreName    string        `json:"store_name"`
	PurchaseDate *string       `json:"purchase_date,omitempty"`
	Total        *float64      `json:"total,omitempty"`
	ItemCount    int           `json:"item_count"`
	CreatedAt    time.Time     `json:"created_at"`
	ImageURL     string        `json:"image_url"`
}

type ReceiptItem struct {
	ID             string
	ReceiptID      string
	RawText        string
	NormalizedName *string
	ProductID      *string
	Quantity       *float64
	UnitPrice      *float64
	Currency       *string
	LineTotal      *float64
}

type Store struct {
	ID      string
	Name    string
	Branch  *string
	Address *string
}

type PriceObservation struct {
	ID         string
	ProductID  string
	StoreID    string
	UnitPrice  float64
	Currency   string
	ObservedAt time.Time
	ReceiptID  string
}

type OCRJob struct {
	ID          string
	ReceiptID   string
	Status      string
	Attempt     int
	LastError   *string
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

type EditableSummary struct {
	ReceiptID     string         `json:"receipt_id"`
	Status        ReceiptStatus  `json:"status"`
	Store         StoreSummary   `json:"store"`
	PurchaseDate  *string        `json:"purchase_date,omitempty"`
	Total         *float64       `json:"total,omitempty"`
	Items         []EditableItem `json:"items"`
	Duplicate     bool           `json:"duplicate"`
	DuplicateOfID *string        `json:"duplicate_of_id,omitempty"`
}

type StoreSummary struct {
	Name    string  `json:"name"`
	Branch  *string `json:"branch,omitempty"`
	Address *string `json:"address,omitempty"`
}

type EditableItem struct {
	RawText   string   `json:"raw_text"`
	Quantity  *float64 `json:"quantity,omitempty"`
	UnitPrice *float64 `json:"unit_price,omitempty"`
	Currency  *string  `json:"currency,omitempty"`
}

type ConfirmPayload struct {
	Store        StoreSummary   `json:"store"`
	PurchaseDate string         `json:"purchase_date"`
	Total        float64        `json:"total"`
	Items        []EditableItem `json:"items"`
}
