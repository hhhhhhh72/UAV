# ---- Build stage ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app ./cmd/api

# ---- Run stage ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

COPY --from=builder /app /app
COPY migrations/ /migrations/

EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app"]
