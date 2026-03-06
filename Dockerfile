# build stage
FROM golang:1.22 AS builder

WORKDIR /app

# dependencies first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# source code
COPY . .

# the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/api


# runtime stage 
FROM alpine:3.19

WORKDIR /app

# copy the compiled binary from builder
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]