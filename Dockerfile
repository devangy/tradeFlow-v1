
# MULTISTAGE build setup to reduce the image size to ~10-20MB
# not exposed to os level vulnerabilities

# STAGE 1 compile
# golang version alpine image lightweight
FROM golang:1.24.7-alpine AS builder

# work dir /app root inside the container
WORKDIR /app

RUN apk add --no-cache ca-certificates


# copy to root
COPY go.mod go.sum ./

# download dependencies
RUN go mod download

# copy files from root directory hostmachine to container /app dir
COPY . .

# create build
RUN CGO_ENABLED=0 GOOS=linux GOMAXPROCS=1 \
    go build -p=1 -o bot-build ./cmd


# STAGE 2 copy only build binary from built image
FROM alpine:latest

#  CA certificate for HTTPS requests
RUN apk --no-cache add ca-certificates

# copy only the final build binary from builder stage
COPY --from=builder /app/bot-build .

#readwrite permission to binary
RUN chmod +x ./bot-build

CMD ["./bot-build"]
