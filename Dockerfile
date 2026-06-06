# --- Build stage ---
FROM golang:1.26-alpine@sha256:f23e8b227fb4493eabe03bede4d5a32d04092da71962f1fb79b5f7d1e6c2a17f AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY css/ css/
COPY internal/ internal/
COPY js/ js/

RUN go run ./cmd/build-assets

ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o direlex ./cmd/server

# --- Final image ---
FROM scratch

WORKDIR /app

COPY data/ data/
COPY public/ public/
COPY --from=builder /app/public/css/ public/css/
COPY --from=builder /app/public/js/ public/js/
COPY --from=builder /app/direlex .

EXPOSE 80

CMD ["./direlex"]
