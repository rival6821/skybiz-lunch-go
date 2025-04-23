# skybiz-lunch-go

## Docker
```
docker build -t skybiz-lunch-go .

docker run --name skybiz-lunch-go-service --rm skybiz-lunch-go
```

## ghcr 저장
```
echo <YOUR_TOKEN> | docker login ghcr.io -u rival6821 --password-stdin

docker build -t ghcr.io/rival6821/skybiz-lunch-go:latest .

docker push ghcr.io/rival6821/skybiz-lunch-go:latest

docker pull ghcr.io/rival6821/skybiz-lunch-go:latest

docker run --name skybiz-lunch-go-service --rm ghcr.io/rival6821/skybiz-lunch-go:latest
```