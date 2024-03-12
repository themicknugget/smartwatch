# Stage 1: Build the Go binary
FROM golang:1.18 as builder
WORKDIR /app

# Copy go mod file
COPY go.mod ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o smartctl-monitor .

# Stage 2: Build environment for compiling smartmontools
FROM alpine:latest as builder2

# Install build dependencies
RUN apk add --no-cache alpine-sdk autoconf automake libtool

# Download and unpack smartmontools
ARG SMARTMONTOOLS_VERSION=7.2
RUN wget https://sourceforge.net/projects/smartmontools/files/smartmontools/${SMARTMONTOOLS_VERSION}/smartmontools-${SMARTMONTOOLS_VERSION}.tar.gz \
    && tar -xzf smartmontools-${SMARTMONTOOLS_VERSION}.tar.gz \
    && rm smartmontools-${SMARTMONTOOLS_VERSION}.tar.gz

# Compile smartmontools statically
WORKDIR /smartmontools-${SMARTMONTOOLS_VERSION}
RUN ./configure CFLAGS="-static" --without-scsi \
    && make && make install

# Stage 2: Final image
FROM scratch

# Copy the smartctl binary
COPY --from=builder /usr/local/sbin/smartctl /usr/local/sbin/smartctl

# Copy the Go program binary
COPY --from=builder2 /app/smartwatch .

# Command to run the executable
CMD ["/smartwatch"]