FROM golang:1.27.1-trixie AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
        libvips-dev \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build BINARY=/out/file-converter

FROM golang:1.27.1-trixie AS dev

RUN apt-get update && apt-get install -y --no-install-recommends \
        libvips-dev \
        libvips42t64 \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

RUN go install github.com/air-verse/air@v1.67.4
RUN go install github.com/go-delve/delve/cmd/dlv@v1.27.2

COPY go.mod go.sum ./
RUN go mod download

EXPOSE 2345 8080
ENTRYPOINT ["air"]

FROM debian:trixie-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        libvips42t64 \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --no-create-home --uid 1000 fileconverter

COPY --from=builder /out/file-converter /usr/local/bin/file-converter

USER fileconverter

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/file-converter"]
