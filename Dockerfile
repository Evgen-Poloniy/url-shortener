FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY configs/ ./configs/
COPY docs/ ./docs/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOGC=100 go build -ldflags="-s -w" -o ./bin/app ./cmd/url-shortener/main.go

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache curl

COPY --from=builder /app/bin/app ./
COPY --from=builder /app/configs ./configs

ENTRYPOINT ["./app"]

CMD ["-storage-type=postgres"]
