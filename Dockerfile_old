# Build stage
FROM golang:1.23 as builder
WORKDIR /src

# Download dependencies using go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy application source
COPY cmd ./cmd
COPY internal ./internal

# Build the server binary
RUN go build -o catalogo-server ./cmd

# Runtime stage
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /src/catalogo-server /usr/local/bin/catalogo-server

# Expose default port
EXPOSE 8080

# Start the server
ENTRYPOINT ["/usr/local/bin/catalogo-server"]
CMD []
