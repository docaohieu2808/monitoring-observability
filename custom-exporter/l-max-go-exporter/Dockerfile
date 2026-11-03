FROM golang:1.27-alpine3.24 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/custom-exporter .

FROM alpine:3.24
RUN adduser -D -u 10001 exporter
COPY --from=builder /out/custom-exporter /usr/local/bin/custom-exporter
USER 10001
EXPOSE 9200
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
  CMD ["wget", "-q", "--spider", "http://127.0.0.1:9200/health"]
ENTRYPOINT ["/usr/local/bin/custom-exporter"]
