from __future__ import annotations

import logging
import os
import threading
from contextlib import asynccontextmanager
from typing import List, Optional, Tuple

import cv2
import numpy as np
import requests
from fastapi import FastAPI, HTTPException
from paddleocr import PaddleOCR
from pydantic import BaseModel

OCR_LANG = os.getenv("OCR_LANG", "en")
OCR_MAX_DIM = int(os.getenv("OCR_MAX_DIM", "2048"))
OCR_DENOISE = int(os.getenv("OCR_DENOISE", "9"))
FETCH_TIMEOUT = 20
LINE_BAND_PX = 10.0
_ocr_lock = threading.Lock()

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("ocr-service")


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Loading PaddleOCR model (lang=%s, max_dim=%s)", OCR_LANG, OCR_MAX_DIM)
    app.state.ocr = PaddleOCR(use_angle_cls=True, lang=OCR_LANG, show_log=False)
    logger.info("PaddleOCR model loaded")
    yield


app = FastAPI(title="ahorrapp-ocr-service", version="0.1.0", lifespan=lifespan)


class ExtractRequest(BaseModel):
    image_ref: str


class ExtractResponse(BaseModel):
    raw_text: str
    lines: List[str]
    confidence: Optional[float] = None


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


def _fetch_image(image_ref: str) -> bytes:
    try:
        resp = requests.get(image_ref, timeout=FETCH_TIMEOUT)
        resp.raise_for_status()
    except requests.RequestException as exc:
        raise HTTPException(status_code=502, detail="fetch_image_failed") from exc
    return resp.content


def _decode_image(data: bytes) -> np.ndarray:
    arr = np.frombuffer(data, np.uint8)
    img = cv2.imdecode(arr, cv2.IMREAD_COLOR)
    if img is None:
        raise HTTPException(status_code=422, detail="decode_image_failed")
    return img


def _preprocess(img: np.ndarray) -> np.ndarray:
    h, w = img.shape[:2]
    longest = max(h, w)
    if longest > OCR_MAX_DIM:
        scale = OCR_MAX_DIM / longest
        new_w = int(w * scale)
        new_h = int(h * scale)
        img = cv2.resize(img, (new_w, new_h), interpolation=cv2.INTER_AREA)
        logger.info("Resized %dx%d -> %dx%d", w, h, new_w, new_h)

    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8, 8))
    gray = clahe.apply(gray)

    if OCR_DENOISE > 0:
        gray = cv2.bilateralFilter(gray, OCR_DENOISE, 75, 75)
        logger.info("Applied bilateral denoise (d=%s)", OCR_DENOISE)

    return cv2.cvtColor(gray, cv2.COLOR_GRAY2BGR)


def _format_result(result) -> Tuple[List[str], Optional[float]]:
    if not result or not result[0]:
        return [], None
    entries = []
    for line in result[0]:
        bbox, (text, score) = line[0], line[1]
        top = min(point[1] for point in bbox)
        left = min(point[0] for point in bbox)
        entries.append((top, left, text.strip(), float(score)))
    entries.sort(key=lambda e: (round(e[0] / LINE_BAND_PX), e[1]))
    lines = [e[2] for e in entries if e[2]]
    scores = [e[3] for e in entries if e[2]]
    confidence = sum(scores) / len(scores) if scores else None
    return lines, confidence


@app.post("/extract", response_model=ExtractResponse)
def extract(payload: ExtractRequest) -> ExtractResponse:
    ocr: PaddleOCR = app.state.ocr
    data = _fetch_image(payload.image_ref)
    img = _decode_image(data)
    img = _preprocess(img)
    with _ocr_lock:
        result = ocr.ocr(img, cls=True)
    lines, confidence = _format_result(result)
    logger.info(
        "Extracted %d lines (confidence=%.3f)",
        len(lines),
        confidence or 0.0,
    )
    return ExtractResponse(
        raw_text="\n".join(lines),
        lines=lines,
        confidence=confidence,
    )
