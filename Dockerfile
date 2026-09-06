# --- Build stage ---
FROM golang:1.27-alpine@sha256:4c9fe60190a2a3350ddc51de80d0224b8a6698d12bdfc999fee45ea9d6c46dbc AS builder

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
