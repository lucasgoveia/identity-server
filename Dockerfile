
# Stage 1: Install dependencies
FROM golang:1.23.2-alpine3.20 AS deps

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Stage 2: Build the application
FROM golang:1.23.2-alpine3.20 AS builder

WORKDIR /app

COPY --from=deps /go/pkg /go/pkg
COPY . .

# Enable them if you need them
# ENV CGO_ENABLED=0
# ENV GOOS=linux

RUN CGO_ENABLED=0 go build -v ./cmd/identity-server/

# Final stage: Run the application
FROM alpine:3.20

WORKDIR /app

# Copy the built application
COPY --from=builder /app/identity-server .

CMD ["./identity-server"]



