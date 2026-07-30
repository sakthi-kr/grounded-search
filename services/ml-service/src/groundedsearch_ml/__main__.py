"""Command-line entry point for the ML service."""

import uvicorn

from groundedsearch_ml.app import create_app
from groundedsearch_ml.config import Settings


def main() -> None:
    """Run the service with Uvicorn."""

    settings = Settings.from_env()
    application = create_app(settings)

    uvicorn.run(
        application,
        host=settings.host,
        port=settings.port,
        access_log=False,
        log_config=None,
    )


if __name__ == "__main__":
    main()
