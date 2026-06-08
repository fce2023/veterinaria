FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .
# Generate go.sum and download dependencies
RUN go mod tidy && go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
