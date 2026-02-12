# --- Build stage ---
FROM golang:1.26-alpine@sha256:d4c4845f5d60c6a974c6000ce58ae079328d03ab7f721a0734277e69905473e5 AS builder

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
