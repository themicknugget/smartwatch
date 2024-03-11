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

# Stage 2: Create a minimal distro
FROM scratch

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/smartctl-monitor .

# Copy smartctl binary and necessary libraries from a temporary alpine image
COPY --from=alpine /usr/sbin/smartctl /usr/sbin/smartctl
COPY --from=alpine /lib/ld-musl-*.so.* /lib/
COPY --from=alpine /usr/lib/libsmartcols.so.1 /usr/lib/
COPY --from=alpine /usr/lib/libudev.so.1 /usr/lib/
COPY --from=alpine /lib/libz.so.1 /lib/

# Command to run the executable
CMD ["./smartctl-monitor"]