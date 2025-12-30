
# MULTISTAGE build setup to reduce the image size to ~10-20MB
# not exposed to os level vulnerabilities

# STAGE 1 compile
# golang version alpine image lightweight
FROM golang:1.24.7-alpine AS builder

# work dir /app root inside the container
WORKDIR /app

# copy to root
COPY cmd/go.mod cmd/go.sum ./

# download dependencies
RUN go mod download

# copy files fromhost machine to container /app dir
COPY . .

# create build
RUN CGO_ENABLED=0 GOOS=linux go build -o bot-build ./cmd/main.go


# STAGE 2 copy only build binary from built image
FROM alpine:latest


#  update certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# copy only the final build binary from builder stage
COPY --from=builder /app/bot-build .

# CMD [ "/app/", touch .env ]


EXPOSE 8080

#

#readwrite permission to binary
RUN chmod +x ./bot-build

CMD ["./bot-build"]
