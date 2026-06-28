// Mock API server for AhorraApp — simulates all /api/v1 endpoints the Flutter app consumes.
// Run with:  go run ./mock-server
// Listens on :8080. Accepts any email/password for register/login.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const addr = ":8080"

var started = time.Now()

type kv = map[string]any

func main() {
	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("/api/v1/auth/register", cors(register))
	mux.HandleFunc("/api/v1/auth/login", cors(login))
	mux.HandleFunc("/api/v1/auth/me", cors(me))

	// Receipts
	mux.HandleFunc("/api/v1/receipts", cors(receiptUpload))  // POST
	mux.HandleFunc("/api/v1/receipts/", cors(receiptDetail)) // GET {id}, POST {id}/confirm

	// Ranking
	mux.HandleFunc("/api/v1/ranking/products/search", cors(productSearch))

	// Profile
	mux.HandleFunc("/api/v1/users/me/points", cors(points))

	// Health
	mux.HandleFunc("/api/v1/health", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, kv{"status": "ok", "since": started.Format(time.RFC3339)})
	}))

	log.Printf("mock-api listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func errJSON(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, kv{"error": msg})
}

// --- Auth --------------------------------------------------------------

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errJSON(w, 405, "method not allowed")
		return
	}
	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errJSON(w, 400, "invalid json")
		return
	}
	if req.Email == "" || len(req.Password) < 8 {
		errJSON(w, 400, "email required and password >= 8 chars")
		return
	}
	name := req.DisplayName
	if name == "" {
		name = strings.Split(req.Email, "@")[0]
	}
	writeJSON(w, 201, kv{
		"token": fakeJWT(req.Email),
		"user": kv{
			"id":           fakeUUID(req.Email),
			"email":        req.Email,
			"display_name": name,
		},
	})
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errJSON(w, 405, "method not allowed")
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errJSON(w, 400, "invalid json")
		return
	}
	if req.Email == "" || req.Password == "" {
		errJSON(w, 401, "invalid credentials")
		return
	}
	// Accept any credentials in mock mode
	writeJSON(w, 200, kv{
		"token": fakeJWT(req.Email),
		"user": kv{
			"id":           fakeUUID(req.Email),
			"email":        req.Email,
			"display_name": strings.Split(req.Email, "@")[0],
		},
	})
}

func me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errJSON(w, 405, "method not allowed")
		return
	}
	email := emailFromAuth(r)
	if email == "" {
		errJSON(w, 401, "invalid or expired token")
		return
	}
	writeJSON(w, 200, kv{
		"id":           fakeUUID(email),
		"email":        email,
		"display_name": strings.Split(email, "@")[0],
	})
}

// --- Receipts ----------------------------------------------------------

// POST /api/v1/receipts  (multipart: image)
func receiptUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errJSON(w, 405, "method not allowed")
		return
	}
	if emailFromAuth(r) == "" {
		errJSON(w, 401, "invalid or expired token")
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		errJSON(w, 400, "invalid multipart form")
		return
	}
	if _, _, err := r.FormFile("image"); err != nil {
		errJSON(w, 400, "image file is required")
		return
	}
	writeJSON(w, 202, kv{
		"receipt_id": "00000000-0000-0000-0000-000000000001",
		"status":     "PENDING",
		"duplicate":  false,
	})
}

// GET /api/v1/receipts/{id}
// POST /api/v1/receipts/{id}/confirm
func receiptDetail(w http.ResponseWriter, r *http.Request) {
	if emailFromAuth(r) == "" {
		errJSON(w, 401, "invalid or expired token")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/receipts/")
	parts := strings.SplitN(path, "/", 2)
	id := parts[0]
	if id == "" {
		errJSON(w, 404, "receipt not found")
		return
	}

	if r.Method == http.MethodGet {
		// Simulate: receipt always returns NEEDS_REVIEW with sample data
		writeJSON(w, 200, kv{
			"receipt_id": id,
			"status":     "NEEDS_REVIEW",
			"store": kv{
				"name":    "Central Madeirense",
				"branch":  "Downtown",
				"address": "Av. Principal #123",
			},
			"purchase_date": "2026-06-25",
			"total":         42.50,
			"items": []kv{
				{
					"raw_text":   "ARROZ 1KG",
					"quantity":   1,
					"unit_price": 2.40,
					"currency":   "USD",
				},
				{
					"raw_text":   "HARINA 1KG",
					"quantity":   2,
					"unit_price": 1.80,
					"currency":   "USD",
				},
			},
		})
		return
	}

	if r.Method == http.MethodPost {
		// /confirm
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		writeJSON(w, 200, kv{"points_earned": 10})
		return
	}

	errJSON(w, 405, "method not allowed")
}

// --- Ranking -----------------------------------------------------------

func productSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errJSON(w, 405, "method not allowed")
		return
	}
	if emailFromAuth(r) == "" {
		errJSON(w, 401, "invalid or expired token")
		return
	}
	q := r.URL.Query().Get("q")
	if q == "" {
		errJSON(w, 400, "query parameter 'q' is required")
		return
	}

	results := []kv{
		{
			"product_id":   "11111111-0000-0000-0000-000000000001",
			"product_name": "Arroz Blanco 1kg",
			"unit":         "kg",
			"stores": []kv{
				{
					"store_id":      "aaaaaaaa-0000-0000-0000-000000000001",
					"store_name":    "Central Madeirense",
					"branch":        "Downtown",
					"average_price": 2.25,
					"currency":      "USD",
					"sample_count":  45,
				},
				{
					"store_id":      "bbbbbbbb-0000-0000-0000-000000000002",
					"store_name":    "SuperMaxi",
					"branch":        "North",
					"average_price": 2.65,
					"currency":      "USD",
					"sample_count":  32,
				},
				{
					"store_id":      "cccccccc-0000-0000-0000-000000000003",
					"store_name":    "Excelsior Gama",
					"branch":        "East",
					"average_price": 2.95,
					"currency":      "USD",
					"sample_count":  20,
				},
			},
		},
		{
			"product_id":   "22222222-0000-0000-0000-000000000002",
			"product_name": "Harina de Maiz 1kg",
			"unit":         "kg",
			"stores": []kv{
				{
					"store_id":      "dddddddd-0000-0000-0000-000000000004",
					"store_name":    "Central Madeirense",
					"branch":        "Downtown",
					"average_price": 1.70,
					"currency":      "USD",
					"sample_count":  28,
				},
				{
					"store_id":      "eeeeeeee-0000-0000-0000-000000000005",
					"store_name":    "Farmatodo",
					"branch":        "Central",
					"average_price": 1.95,
					"currency":      "USD",
					"sample_count":  15,
				},
			},
		},
	}

	writeJSON(w, 200, kv{"results": results})
}

// --- Profile -----------------------------------------------------------

func points(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errJSON(w, 405, "method not allowed")
		return
	}
	if emailFromAuth(r) == "" {
		errJSON(w, 401, "invalid or expired token")
		return
	}
	writeJSON(w, 200, kv{
		"total_points": 350,
		"recent_transactions": []kv{
			{
				"id":         "tx-001",
				"points":     10,
				"reason":     "Receipt confirmed",
				"created_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			},
			{
				"id":         "tx-002",
				"points":     10,
				"reason":     "Receipt confirmed",
				"created_at": time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			},
		},
	})
}

// --- Helpers -----------------------------------------------------------

// fakeJWT: returns a header-prefixed dummy token (NOT a real JWT, but the
// Flutter app only stores and echoes it back; the mock decodes the email).
func fakeJWT(email string) string {
	// Encoded as: mock.<email>
	return "mock." + email
}

func emailFromAuth(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	tok := strings.TrimPrefix(auth, "Bearer ")
	if !strings.HasPrefix(tok, "mock.") {
		return ""
	}
	return strings.TrimPrefix(tok, "mock.")
}

func fakeUUID(seed string) string {
	// Deterministic-ish UUIDv4 format based on email
	h := uint32(0)
	for _, c := range seed {
		h = h*31 + uint32(c)
	}
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", h)
}
