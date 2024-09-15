# Dockerfile - for server

FROM golang:1.23-alpine

WORKDIR /app

# Copy source code
COPY . .

# Download Go modules
RUN go mod download

# Build Server
RUN go build -v -o websocket-server .

# Debugging step: Check if the binary was created
RUN ls -la /app

# Expose port 8081
EXPOSE 8081

# Start the server
CMD ["./websocket-server"]