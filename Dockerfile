FROM golang:1.22-alpine AS builder

# Set CGO_ENABLED for go-sqlite3
ENV CGO_ENABLED=1

# Install build tools needed for CGO and SQLite
# build-base includes gcc, musl-dev, etc.
RUN apk add --no-cache build-base gcc

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Compile the application
# Ensure main.go, database.go (and any other necessary .go files) are in the current context
RUN go build -o /app/api_server .

# Start a new stage from a minimal base image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the pre-compiled binary from the builder stage
COPY --from=builder /app/api_server /app/api_server

# The events.db will be mounted via docker-compose into /app/data
# Create the data directory in case it's not created by the volume mount
RUN mkdir -p /app/data

# Expose port 3000 (or whatever port the app listens on)
EXPOSE 3000

# Command to run the application
CMD ["/app/api_server"] 