
FROM golang:1.27-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN --network=host wget -T 15 -S -O /dev/null https://proxy.golang.org/filippo.io/edwards25519/@v/v1.2.0.mod
RUN --network=host printf '%s\n' \
'package main' \
'import ("fmt"; "net")' \
'func main() { ips, err := net.LookupIP("proxy.golang.org"); fmt.Println("Resolved IPs:", ips, "Error:", err) }' \
> /tmp/dns.go && go run /tmp/dns.go
RUN --network=host printf '%s\n' \
'package main' \
'import ("crypto/tls"; "fmt"; "net"; "time")' \
'func main() { conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10*time.Second}, "tcp4", "142.251.13.141:443", &tls.Config{ServerName: "proxy.golang.org"}); fmt.Println("IPv4 TLS error:", err); if conn != nil { conn.Close() } }' \
> /tmp/ipv4.go && go run /tmp/ipv4.go
RUN --network=host GODEBUG=http2client=0 GOMAXPROCS=2 GOFLAGS="-p=1" go mod download

COPY . .

RUN CGO_ENABLED=0 go build -mod=mod -o /out/receiver ./cmd/receiver
RUN CGO_ENABLED=0 go build -mod=mod -o /out/checker ./cmd/checker

FROM alpine:3.21

RUN adduser -D -H appuser

WORKDIR /app

COPY --from=build /out/receiver /app/receiver
COPY --from=build /out/checker /app/checker

USER appuser

EXPOSE 8080 50051
