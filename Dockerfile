# Use the official Golang image as a base
FROM golang:1.22-bullseye

# Install dependencies for AF_XDP and benchmarking
RUN apt-get update && apt-get install -y \
    libbpf-dev \
    clang \
    llvm \
    libelf-dev \
    iproute2 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy dependency files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the mohawk-node
RUN go build -o mohawk-node ./cmd/mohawk-node

# Default command: run the standard test suite
CMD ["go", "test", "./...", "-v"]
