
env:
	@docker-compose up -d

dep:
	@dep ensure -vendor-only
	@go get github.com/golang/lint

test:
	@go test

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/fetch-publication-pages ./cmd/fetch-publication-pages/
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/parse-publication-pages ./cmd/parse-publication-pages/
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/push-publications ./cmd/push-publications/
