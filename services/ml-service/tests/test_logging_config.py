"""Tests for structured logging."""

import json
import logging

from groundedsearch_ml.logging_config import JSONFormatter, configure_logging


def test_json_formatter_includes_structured_fields() -> None:
    formatter = JSONFormatter()
    record = logging.LogRecord(
        name="groundedsearch_ml.test",
        level=logging.INFO,
        pathname=__file__,
        lineno=10,
        msg="request complete",
        args=(),
        exc_info=None,
    )
    record.request_id = "request-123"
    record.status = 200

    payload = json.loads(formatter.format(record))

    assert payload["level"] == "INFO"
    assert payload["message"] == "request complete"
    assert payload["request_id"] == "request-123"
    assert payload["status"] == 200
    assert payload["timestamp"]


def test_json_formatter_includes_exception() -> None:
    formatter = JSONFormatter()

    try:
        raise RuntimeError("test failure")
    except RuntimeError:
        record = logging.LogRecord(
            name="groundedsearch_ml.test",
            level=logging.ERROR,
            pathname=__file__,
            lineno=30,
            msg="failed",
            args=(),
            exc_info=__import__("sys").exc_info(),
        )

    payload = json.loads(formatter.format(record))

    assert "RuntimeError: test failure" in payload["exception"]


def test_configure_logging_is_idempotent() -> None:
    root_logger = logging.getLogger()
    original_handlers = list(root_logger.handlers)
    original_level = root_logger.level

    try:
        root_logger.handlers.clear()
        configure_logging(logging.DEBUG)
        configure_logging(logging.INFO)

        groundedsearch_handlers = [
            handler
            for handler in root_logger.handlers
            if handler.get_name() == "groundedsearch-json"
        ]

        assert len(groundedsearch_handlers) == 1
        assert root_logger.level == logging.INFO
    finally:
        root_logger.handlers[:] = original_handlers
        root_logger.setLevel(original_level)
