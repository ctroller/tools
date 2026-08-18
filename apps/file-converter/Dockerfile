FROM golang:1.26-trixie AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
        libvips-dev \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build BINARY=/out/file-converter

FROM debian:trixie-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        libvips42t64 \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --no-create-home --uid 1000 fileconverter

COPY --from=builder /out/file-converter /usr/local/bin/file-converter

USER fileconverter

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/file-converter"]
