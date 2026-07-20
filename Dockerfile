# --- Build stage ---
FROM golang:1.26-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY css/ css/
COPY internal/ internal/
COPY js/ js/

RUN go run ./cmd/build-assets

ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o direlex ./cmd/server

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
