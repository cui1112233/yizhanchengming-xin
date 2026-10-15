"""Resident OCR worker for the Windows giant-material executor.

The process speaks newline-delimited JSON over stdin/stdout.  It deliberately
keeps the process and loaded OCR runtime alive between jobs; only the current
job's temporary decoder state is released after an ``idle`` event.
"""

from __future__ import annotations

import difflib
import json
import math
import os
import queue
import re
import sys
import threading
import time
from dataclasses import dataclass
from typing import Any, Callable, TextIO
from urllib.parse import urlparse


MAX_OUTPUT_BYTES = 2 * 1024 * 1024
ALLOWED_VIDEO_HOSTS = {"material.hnqingyuwen.top", "mlzr-material.hnqingyuwen.top", "ml-material.hnqingyuwen.top"}
HAN_OR_TEXT = re.compile(r"[\u3400-\u9fffA-Za-z0-9]")


class WorkerError(Exception):
    def __init__(self, code: str, message: str | None = None):
        super().__init__(message or code)
        self.code = code


@dataclass
class _Job:
    request: dict[str, Any]


class Worker:
    def __init__(self, output: TextIO, extractor: Callable | None = None):
        self.output = output
        self.extractor = extractor or self._extract_video
        self.model_dir: str | None = None
        self._jobs: queue.Queue[_Job | None] = queue.Queue()
        self._stop = threading.Event()
        self._idle = threading.Event()
        self._idle.set()
        self._current_lock = threading.Lock()
        self._current_job_id: str | None = None
        self._cancel_event: threading.Event | None = None
        self._output_lock = threading.Lock()
        self._job_thread = threading.Thread(target=self._job_loop, name="giant-ocr-worker", daemon=True)
        self._job_thread.start()

    def run(self, input_stream: TextIO) -> None:
        for raw in input_stream:
            raw = raw.strip()
            if not raw:
                continue
            try:
                command = json.loads(raw)
            except json.JSONDecodeError:
                self._emit({"type": "failed", "code": "WORKER_INVALID_JSON", "message": "invalid command"})
                continue
            if not isinstance(command, dict):
                self._emit({"type": "failed", "code": "WORKER_INVALID_COMMAND", "message": "command must be an object"})
                continue
            self.handle_command(command)
            if command.get("type") == "shutdown":
                break
        self.shutdown()

    def handle_command(self, command: dict[str, Any]) -> None:
        command_type = command.get("type")
        if command_type == "ensure_model":
            model_dir = str(command.get("modelDir") or "").strip()
            if not model_dir:
                self._emit({"type": "failed", "code": "MODEL_NOT_READY", "message": "modelDir is required"})
                return
            self.model_dir = model_dir
            self._emit({"type": "ready", "modelVersion": command.get("modelVersion", ""), "modelDir": model_dir})
            return
        if command_type == "extract":
            self._enqueue_extract(command)
            return
        if command_type == "cancel":
            job_id = str(command.get("jobId") or "").strip()
            with self._current_lock:
                if job_id and job_id == self._current_job_id and self._cancel_event is not None:
                    self._cancel_event.set()
            return
        if command_type == "shutdown":
            self.shutdown()
            return
        self._emit({"type": "failed", "code": "WORKER_INVALID_COMMAND", "message": "unsupported command"})

    def wait_until_idle(self, timeout: float = 10.0) -> bool:
        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            if self._idle.is_set() and self._jobs.empty():
                with self._current_lock:
                    if self._current_job_id is None:
                        return True
            time.sleep(0.005)
        return False

    def shutdown(self) -> None:
        if self._stop.is_set():
            return
        self._stop.set()
        with self._current_lock:
            if self._cancel_event is not None:
                self._cancel_event.set()
        while True:
            try:
                self._jobs.get_nowait()
            except queue.Empty:
                break
        self._jobs.put(None)
        if threading.current_thread() is not self._job_thread:
            self._job_thread.join(timeout=5)

    def _enqueue_extract(self, request: dict[str, Any]) -> None:
        try:
            checked = validate_extract_request(request)
        except WorkerError as error:
            self._emit_failed(request.get("jobId"), error)
            self._emit_idle(request.get("jobId"))
            return
        self._jobs.put(_Job(checked))

    def _job_loop(self) -> None:
        while True:
            item = self._jobs.get()
            if item is None:
                return
            request = item.request
            job_id = request["jobId"]
            cancel_event = threading.Event()
            with self._current_lock:
                self._current_job_id = job_id
                self._cancel_event = cancel_event
            self._idle.clear()
            try:
                result = self.extractor(request, lambda completed, total, percent: self._emit_progress(job_id, completed, total, percent), cancel_event)
                if cancel_event.is_set():
                    raise WorkerError("OCR_CANCELLED")
                text = str(result.get("text") or "")
                if not text.strip():
                    raise WorkerError("OCR_NO_TEXT")
                if len(text.encode("utf-8")) > MAX_OUTPUT_BYTES:
                    raise WorkerError("OCR_OUTPUT_TOO_LARGE")
                self._emit({"type": "complete", "jobId": job_id, **{key: value for key, value in result.items() if key != "text"}, "text": text})
            except WorkerError as error:
                self._emit_failed(job_id, error)
            except Exception as error:  # noqa: BLE001 - process boundary must remain alive
                self._emit_failed(job_id, WorkerError("OCR_EXECUTION_FAILED", str(error)))
            finally:
                with self._current_lock:
                    self._current_job_id = None
                    self._cancel_event = None
                self._idle.set()
                self._emit_idle(job_id)

    def _emit_progress(self, job_id: str, completed: int, total: int, percent: int) -> None:
        self._emit({"type": "progress", "jobId": job_id, "completed": completed, "total": total, "percent": percent})

    def _emit_failed(self, job_id: Any, error: WorkerError) -> None:
        self._emit({"type": "failed", "jobId": job_id or "", "code": error.code, "message": str(error)[:512]})

    def _emit_idle(self, job_id: Any) -> None:
        self._emit({"type": "idle", "jobId": job_id or ""})

    def _emit(self, event: dict[str, Any]) -> None:
        with self._output_lock:
            self.output.write(json.dumps(event, ensure_ascii=False, separators=(",", ":")) + "\n")
            self.output.flush()

    def _extract_video(self, request: dict[str, Any], progress: Callable, cancel_event: threading.Event) -> dict[str, Any]:
        """Decode one frame per second and OCR the text panel.

        PaddleOCR/OpenCV are imported only when a real job starts, so the
        lightweight installer can boot before the first-use runtime download.
        """
        try:
            import cv2  # type: ignore
            from paddleocr import PaddleOCR  # type: ignore
        except ImportError as error:
            raise WorkerError("OCR_RUNTIME_NOT_READY", str(error)) from error

        capture = cv2.VideoCapture(request["videoUrl"])
        if not capture.isOpened():
            raise WorkerError("OCR_VIDEO_READ_FAILED")
        try:
            fps = capture.get(cv2.CAP_PROP_FPS) or 30.0
            duration = float(request["durationSeconds"])
            total = max(1, int(math.ceil(duration)))
            ocr = PaddleOCR(use_angle_cls=True, lang="ch", show_log=False)
            body = ""
            previous = ""
            frames = 0
            duplicates = 0
            for second in range(total):
                if cancel_event.is_set():
                    raise WorkerError("OCR_CANCELLED")
                capture.set(cv2.CAP_PROP_POS_MSEC, second * 1000)
                ok, frame = capture.read()
                if not ok:
                    continue
                frames += 1
                text = ocr_frame_text(ocr, frame)
                if text:
                    merged, skipped = append_scroll_text(body, previous, text)
                    body = merged
                    previous = text
                    duplicates += int(skipped)
                progress(second + 1, total, int((second + 1) * 100 / total))
            body = clean_text(body)
            if not body:
                raise WorkerError("OCR_NO_TEXT")
            return {"text": body, "characters": len(re.sub(r"\s", "", body)), "frames": frames, "duplicates": duplicates}
        finally:
            capture.release()


def validate_extract_request(request: dict[str, Any]) -> dict[str, Any]:
    job_id = str(request.get("jobId") or "").strip()
    raw_url = str(request.get("videoUrl") or "").strip()
    parsed = urlparse(raw_url)
    duration = request.get("durationSeconds")
    if not job_id:
        raise WorkerError("OCR_INVALID_JOB")
    if parsed.scheme != "https" or parsed.hostname not in ALLOWED_VIDEO_HOSTS or parsed.username or parsed.password or parsed.port:
        raise WorkerError("OCR_VIDEO_NOT_ALLOWED")
    try:
        duration_value = float(duration)
    except (TypeError, ValueError):
        raise WorkerError("OCR_DURATION_NOT_SUPPORTED") from None
    if not math.isfinite(duration_value) or duration_value <= 0 or duration_value > 1800:
        raise WorkerError("OCR_DURATION_NOT_SUPPORTED")
    return {**request, "jobId": job_id, "videoUrl": raw_url, "durationSeconds": duration_value}


def clean_text(value: str) -> str:
    lines = []
    for line in value.replace("\r\n", "\n").replace("\r", "\n").split("\n"):
        line = line.strip()
        if not line or HAN_OR_TEXT.search(line):
            lines.append(line)
    return "\n".join(lines).strip()


def append_scroll_text(body: str, previous: str, current: str) -> tuple[str, bool]:
    current = current.strip()
    if not current:
        return body, True
    if not body:
        return current, False
    normalized_current = normalize_characters(current)
    tail = normalize_characters(body[-2500:])
    if normalized_current and normalized_current in tail:
        return body, True
    previous_norm = normalize_characters(previous)[-600:]
    next_norm, next_ends = normalized_with_ends(current)
    next_norm = next_norm[:600]
    match = difflib.SequenceMatcher(None, previous_norm, next_norm, autojunk=False).find_longest_match(0, len(previous_norm), 0, len(next_norm))
    if match.size >= 20 and match.b == 0:
        extra = current[next_ends[match.size - 1]:] if match.size <= len(next_ends) else ""
        if extra:
            return remove_duplicate_boundary_symbols(body, extra) + extra, False
        return body, True
    return body + "\n\n" + current, False


def normalize_characters(value: str) -> str:
    return "".join(character for character in value if HAN_OR_TEXT.match(character))


def normalized_with_ends(value: str) -> tuple[str, list[int]]:
    characters = []
    ends = []
    offset = 0
    for character in value:
        offset += 1
        if HAN_OR_TEXT.match(character):
            characters.append(character)
            ends.append(offset)
    return "".join(characters), ends


def remove_duplicate_boundary_symbols(body: str, extra: str) -> str:
    max_size = min(8, len(body), len(extra))
    for size in range(max_size, 0, -1):
        suffix = body[-size:]
        prefix = extra[:size]
        if suffix == prefix and suffix.strip() and all(not HAN_OR_TEXT.match(character) for character in suffix):
            return body[:-size]
    return body


def ocr_frame_text(ocr: Any, frame: Any) -> str:
    result = ocr.ocr(frame, cls=True)
    lines: list[str] = []
    for page in result or []:
        for item in page or []:
            if len(item) < 2:
                continue
            text = str(item[1][0]).strip()
            if text and HAN_OR_TEXT.search(text):
                lines.append(text)
    return "\n".join(lines)


def main() -> None:
    worker = Worker(sys.stdout)
    worker.run(sys.stdin)


if __name__ == "__main__":
    main()
