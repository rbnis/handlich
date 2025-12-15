FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./

# Copy the entire project structure
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the application
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -ldflags="-s -w" -o handlich ./cmd/handlich/main.go

FROM scratch

EXPOSE 8080/tcp

# Copy CA certificates from alpine for SSL/TLS verification
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

COPY --from=build /app/handlich .

ENTRYPOINT [ "./handlich" ]
