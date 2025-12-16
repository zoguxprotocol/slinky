FROM golang:1.25.1-trixie
LABEL org.opencontainers.image.source="https://github.com/zoguxprotocol/slinky"

WORKDIR /src/slinky

RUN apt-get update \
    && apt-get install -y --no-install-recommends jq ca-certificates make git curl bash dasel \
    && rm -rf /var/lib/apt/lists/*

ENV GOTOOLCHAIN=auto
