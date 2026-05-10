FROM golang:1.26.3-alpine AS builder
WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o gopher-pinger ./cmd/pinger/main.go

FROM alpine:3.19

RUN adduser -D appuser
USER appuser

WORKDIR /app
COPY --from=builder /app/gopher-pinger .

CMD [ "./gopher-pinger" ]