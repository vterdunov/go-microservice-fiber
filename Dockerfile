# Build stage
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# Final stage
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /server /server

EXPOSE 3000

ENTRYPOINT ["/server"]
