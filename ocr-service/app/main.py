from __future__ import annotations

from typing import List, Optional

import requests
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(title="ahorrapp-ocr-service", version="0.1.0")


class ExtractRequest(BaseModel):
    image_ref: str


class ExtractResponse(BaseModel):
    raw_text: str
    lines: List[str]
    confidence: Optional[float] = None


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/extract", response_model=ExtractResponse)
def extract(payload: ExtractRequest) -> ExtractResponse:
    try:
        resp = requests.get(payload.image_ref, timeout=20)
        resp.raise_for_status()
    except requests.RequestException as exc:
        raise HTTPException(status_code=502, detail="fetch_image_failed") from exc

    # Minimal MVP extraction path. In production, integrate PaddleOCR
    # processing of image bytes and map output to raw_text/lines/confidence.
    sample_lines = [
        "SUPERMARKET CENTRAL",
        "DATE: 2026-06-24",
        "LECHE 1 x 3.50 USD",
    ]
    return ExtractResponse(
        raw_text="\n".join(sample_lines),
        lines=sample_lines,
        confidence=0.9,
    )
