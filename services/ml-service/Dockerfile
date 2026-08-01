# syntax=docker/dockerfile:1.7

FROM python:3.10.20-slim-bookworm AS runtime

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    PIP_NO_CACHE_DIR=1

WORKDIR /app

RUN groupadd --system groundedsearch \
    && useradd \
      --system \
      --gid groundedsearch \
      --home-dir /nonexistent \
      --shell /usr/sbin/nologin \
      groundedsearch

COPY services/ml-service/pyproject.toml \
     services/ml-service/README.md \
     ./
COPY services/ml-service/src ./src

RUN python -m pip install --upgrade "pip==26.2" \
    && python -m pip install .

USER groundedsearch

EXPOSE 8090

ENTRYPOINT ["python", "-m", "groundedsearch_ml"]
