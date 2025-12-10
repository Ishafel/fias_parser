# Build the streaming GAR XML parser
FROM golang:1.21 AS builder
WORKDIR /app

# Download dependencies first for better layer caching
COPY go.mod ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build statically linked binary for a small runtime image
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/fias-parser ./...

FROM debian:bookworm-slim
WORKDIR /
COPY --from=builder /bin/fias-parser /usr/local/bin/
COPY gar_schemas /gar_schemas

ENTRYPOINT ["/usr/local/bin/fias-parser"]
CMD ["--help"]
