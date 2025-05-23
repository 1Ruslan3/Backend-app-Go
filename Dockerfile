FROM golang:1.24.2 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o app ./cmd/main.go

FROM gcr.io/distroless/base-debian10

WORKDIR /app
COPY --from=builder /app/app .

EXPOSE 8080
ENTRYPOINT ["/app/app"]
