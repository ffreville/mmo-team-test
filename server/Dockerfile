FROM golang:1.21-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /build/bin/ /app/bin/
COPY --from=builder /build/configs/ /app/configs/

EXPOSE 8080 8081 8082

CMD ["./bin/game-server"]
