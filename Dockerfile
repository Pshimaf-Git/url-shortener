FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /app/bin/url-shortener ./api/cmd/app

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app


COPY --from=builder /app/bin/url-shortener .

COPY api/configs/config.yaml /app/configs/config.yaml

EXPOSE 8000

CMD [ "./url-shortener" ]