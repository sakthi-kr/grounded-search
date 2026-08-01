"""Tests for the ML service command-line entry point."""

from typing import Any

from groundedsearch_ml import __main__
from groundedsearch_ml.config import Settings


def test_main_starts_uvicorn(monkeypatch: Any) -> None:
    settings = Settings(
        host="127.0.0.1",
        port=9123,
        log_level="CRITICAL",
        environment="test",
    )
    captured: dict[str, object] = {}

    monkeypatch.setattr(
        __main__.Settings,
        "from_env",
        classmethod(lambda cls: settings),
    )

    def fake_run(
        application: object,
        **kwargs: object,
    ) -> None:
        captured["application"] = application
        captured.update(kwargs)

    monkeypatch.setattr(__main__.uvicorn, "run", fake_run)

    __main__.main()

    assert captured["host"] == "127.0.0.1"
    assert captured["port"] == 9123
    assert captured["access_log"] is False
    assert captured["log_config"] is None
    assert captured["application"] is not None
