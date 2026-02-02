FROM golang:1.25.5-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o spot-service ./cmd/main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/spot-service /app/spot-service
COPY configs /app/configs

EXPOSE 50052

ENTRYPOINT ["/app/spot-service"]