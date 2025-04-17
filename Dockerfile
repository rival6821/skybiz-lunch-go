# 빌드 스테이지
FROM golang:1.24-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app .

# 실행 스테이지 (최소 이미지)
FROM debian:bookworm-slim

WORKDIR /app

# 필요한 패키지 설치 (예: ca-certificates)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app .

CMD ["./app"]