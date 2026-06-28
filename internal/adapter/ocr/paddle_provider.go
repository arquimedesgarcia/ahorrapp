package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ahorrapp/internal/domain/ports"
)

type PaddleOCRProvider struct {
	baseURL string
	http    *http.Client
}

func NewPaddleOCRProvider(baseURL string) *PaddleOCRProvider {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "http://localhost:8081"
	}
	return &PaddleOCRProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *PaddleOCRProvider) Extract(ctx context.Context, imageRef string) (ports.RawOCRResult, error) {
	payload := map[string]string{"image_ref": imageRef}
	body, err := json.Marshal(payload)
	if err != nil {
		return ports.RawOCRResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/extract", bytes.NewReader(body))
	if err != nil {
		return ports.RawOCRResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return ports.RawOCRResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return ports.RawOCRResult{}, fmt.Errorf("ocr service returned status %d", resp.StatusCode)
	}

	var out struct {
		RawText    string   `json:"raw_text"`
		Lines      []string `json:"lines"`
		Confidence *float64 `json:"confidence"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ports.RawOCRResult{}, err
	}

	return ports.RawOCRResult{RawText: out.RawText, Lines: out.Lines, Confidence: out.Confidence}, nil
}
