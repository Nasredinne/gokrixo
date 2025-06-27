# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go app
RUN go build -o krixo

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy the compiled binary from builder
COPY --from=builder /app/krixo .

# Run the Go app
EXPOSE 3000
CMD ["./krixo"]