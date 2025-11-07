# Stage 1: Build the Go application
FROM golang:1.22-bullseye AS builder

RUN apt-get install -y  git gcc g++ libc-dev curl
# Set the Current Working Directory inside the container
WORKDIR /app


ENV GOPROXY=https://goproxy.cn,direct



# Set the environment variables
RUN go env -w GOPROXY=$GO_PROXY
RUN go env -w GOMODCACHE=/go/pkg/mod

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download -x
# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o main .


# Stage 2: Create a small image
FROM debian:bullseye-slim as production



# Set the timezone to Singapore time
RUN apt-get update && apt-get install -y tzdata && \
	    ln -fs /usr/share/zoneinfo/Asia/Singapore /etc/localtime && \
	    dpkg-reconfigure -f noninteractive tzdata && \
	    apt-get clean && rm -rf /var/lib/apt/lists/*
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set the Current Working Directory inside the container
WORKDIR /app

RUN mkdir /app/conf


# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

ADD ./config/config.yaml /app/conf/config.yaml
# Command to run the executable
CMD ["./main", "-config=/app/conf/config.yaml"]
