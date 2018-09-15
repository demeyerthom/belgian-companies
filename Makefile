
env:
	@docker-compose --file deployments/docker-compose.yaml up -d

dep:
	@dep ensure -vendor-only
	@go get github.com/golang/lint

test:
	@go test
