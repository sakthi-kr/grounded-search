"""Structured JSON logging for the ML service."""

from __future__ import annotations

import json
import logging
import sys
from datetime import datetime, timezone
from typing import Any


class JSONFormatter(logging.Formatter):
    """Format log records as compact JSON objects."""

    _extra_fields = (
        "request_id",
        "method",
        "path",
        "status",
        "duration_ms",
    )

    def format(self, record: logging.LogRecord) -> str:
        payload: dict[str, Any] = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
        }

        for field in self._extra_fields:
            value = getattr(record, field, None)
            if value is not None:
                payload[field] = value

        if record.exc_info is not None:
            payload["exception"] = self.formatException(record.exc_info)

        return json.dumps(
            payload,
            ensure_ascii=False,
            separators=(",", ":"),
        )


def configure_logging(log_level: int) -> None:
    """Configure the process root logger once."""

    root_logger = logging.getLogger()
    root_logger.setLevel(log_level)

    if any(
        handler.get_name() == "groundedsearch-json" for handler in root_logger.handlers
    ):
        return

    handler = logging.StreamHandler(sys.stdout)
    handler.set_name("groundedsearch-json")
    handler.setFormatter(JSONFormatter())
    root_logger.addHandler(handler)
