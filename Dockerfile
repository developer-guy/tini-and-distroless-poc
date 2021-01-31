# Specify base image
FROM golang:1.15.7-alpine as builder

# Specify working directory
WORKDIR /app

# Add Tini init-system
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-static /tini
RUN chmod +x /tini

# Define environment variables for go build time
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Efficient cache usage
COPY go.mod go.sum ./
RUN go mod download

# Copy everything from host to the image
COPY . .

# Build statically compiled binary
RUN go build -o hello-world

FROM gcr.io/distroless/static-debian10

COPY --from=builder /app/hello-world ./
COPY --from=builder /tini /tini

ENTRYPOINT ["/tini", "--"]
CMD ["./hello-world"]
