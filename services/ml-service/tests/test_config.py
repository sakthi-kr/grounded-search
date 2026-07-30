"""Tests for environment-backed configuration."""

import logging

import pytest

from groundedsearch_ml.config import Settings


def test_defaults() -> None:
    settings = Settings.from_env({})

    assert settings.host == "0.0.0.0"
    assert settings.port == 8090
    assert settings.log_level == "INFO"
    assert settings.environment == "development"
    assert settings.numeric_log_level == logging.INFO


def test_overrides() -> None:
    settings = Settings.from_env(
        {
            "ML_SERVICE_HOST": "127.0.0.1",
            "ML_SERVICE_PORT": "9090",
            "LOG_LEVEL": "debug",
            "GROUNDEDSEARCH_ENV": "test",
        }
    )

    assert settings.host == "127.0.0.1"
    assert settings.port == 9090
    assert settings.log_level == "DEBUG"
    assert settings.environment == "test"
    assert settings.numeric_log_level == logging.DEBUG


@pytest.mark.parametrize(
    ("environment", "message"),
    [
        (
            {"ML_SERVICE_PORT": "not-a-number"},
            "ML_SERVICE_PORT must be an integer",
        ),
        (
            {"ML_SERVICE_PORT": "70000"},
            "ML_SERVICE_PORT must be between 1 and 65535",
        ),
        (
            {"LOG_LEVEL": "verbose"},
            "LOG_LEVEL must be one of",
        ),
    ],
)
def test_invalid_values(
    environment: dict[str, str],
    message: str,
) -> None:
    with pytest.raises(ValueError, match=message):
        Settings.from_env(environment)


def test_blank_values_use_defaults() -> None:
    settings = Settings.from_env(
        {
            "ML_SERVICE_HOST": " ",
            "ML_SERVICE_PORT": " ",
            "LOG_LEVEL": " ",
            "GROUNDEDSEARCH_ENV": " ",
        }
    )

    assert settings.host == "0.0.0.0"
    assert settings.port == 8090
    assert settings.log_level == "INFO"
    assert settings.environment == "development"
