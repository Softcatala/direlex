# --- Build stage ---
FROM golang:1.26-alpine@sha256:7a3e50096189ad57c9f9f865e7e4aa8585ed1585248513dc5cda498e2f41812c AS builder

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
