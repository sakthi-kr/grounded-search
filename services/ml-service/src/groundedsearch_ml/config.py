"""Environment-backed configuration for the ML service."""

from __future__ import annotations

import logging
import os
from collections.abc import Mapping
from dataclasses import dataclass

_DEFAULT_HOST = "0.0.0.0"
_DEFAULT_PORT = 8090
_DEFAULT_LOG_LEVEL = "INFO"
_DEFAULT_ENVIRONMENT = "development"
_LOG_LEVEL_VALUES = {
    "CRITICAL": logging.CRITICAL,
    "ERROR": logging.ERROR,
    "WARNING": logging.WARNING,
    "INFO": logging.INFO,
    "DEBUG": logging.DEBUG,
}


@dataclass(frozen=True, slots=True)
class Settings:
    """Validated runtime settings."""

    host: str = _DEFAULT_HOST
    port: int = _DEFAULT_PORT
    log_level: str = _DEFAULT_LOG_LEVEL
    environment: str = _DEFAULT_ENVIRONMENT

    @classmethod
    def from_env(
        cls,
        environ: Mapping[str, str] | None = None,
    ) -> Settings:
        """Load settings from environment variables."""

        source = os.environ if environ is None else environ

        host = source.get("ML_SERVICE_HOST", _DEFAULT_HOST).strip()
        if not host:
            host = _DEFAULT_HOST

        port = _parse_port(source.get("ML_SERVICE_PORT"))

        log_level = (
            source.get(
                "LOG_LEVEL",
                _DEFAULT_LOG_LEVEL,
            )
            .strip()
            .upper()
        )
        if not log_level:
            log_level = _DEFAULT_LOG_LEVEL
        if log_level not in _LOG_LEVEL_VALUES:
            allowed = ", ".join(sorted(_LOG_LEVEL_VALUES))
            raise ValueError(f"LOG_LEVEL must be one of: {allowed}")

        environment = source.get(
            "GROUNDEDSEARCH_ENV",
            _DEFAULT_ENVIRONMENT,
        ).strip()
        if not environment:
            environment = _DEFAULT_ENVIRONMENT

        return cls(
            host=host,
            port=port,
            log_level=log_level,
            environment=environment,
        )

    @property
    def numeric_log_level(self) -> int:
        """Return the standard-library numeric log level."""

        return _LOG_LEVEL_VALUES[self.log_level]


def _parse_port(raw: str | None) -> int:
    if raw is None or not raw.strip():
        return _DEFAULT_PORT

    try:
        port = int(raw.strip())
    except ValueError as error:
        raise ValueError("ML_SERVICE_PORT must be an integer") from error

    if not 1 <= port <= 65535:
        raise ValueError("ML_SERVICE_PORT must be between 1 and 65535")

    return port
