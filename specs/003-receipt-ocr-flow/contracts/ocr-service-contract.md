# Contract: OCR Service (Internal)

Service is deployed independently and consumed by `PaddleOCRProvider` adapter.

## POST `/extract`

### Request

```json
{
  "image_ref": "http://minio:9000/receipts/abc.jpg"
}
```

### Success (`200`)

```json
{
  "raw_text": "SUPERMARKET CENTRAL\nDATE: 2026-06-21\nARROZ 1KG 1 x 2.40 USD",
  "lines": [
    "SUPERMARKET CENTRAL",
    "DATE: 2026-06-21",
    "ARROZ 1KG 1 x 2.40 USD"
  ],
  "confidence": 0.91
}
```

### Failure (`502`)

```json
{
  "error": "extract_failed"
}
```

## GET `/health`

### Success (`200`)

```json
{
  "status": "ok"
}
```
