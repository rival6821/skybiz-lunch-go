FROM golang:1.24

# libwebp 및 필요한 패키지 설치
RUN apt-get update && apt-get install -y

WORKDIR /app

# Go 모듈 복사 및 의존성 설치
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# 애플리케이션 빌드
RUN go build -o app .

# 실행할 명령어
CMD ["./app"]