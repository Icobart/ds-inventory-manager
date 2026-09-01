# Build the go binary
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the node daemon. 
RUN CGO_ENABLED=0 GOOS=linux go build -a -o warehouse-node ./cmd/node

# Create the minimal production image
FROM alpine:3.24

WORKDIR /app

COPY --from=builder /app/warehouse-node .

EXPOSE 50051

ENTRYPOINT ["./warehouse-node"]