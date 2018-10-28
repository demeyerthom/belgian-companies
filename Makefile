
env: ## Create testing environment
	@docker-compose up -d

dependencies: ## Ensure dependencies
	@go get -u golang.org/x/lint/golint
	@dep ensure -vendor-only

lint:
	@golint pkg/* cmd/*

test: ## Run tests
	@go test ./...

clean: ## Clean project files
	rm -rfd vendor
	rm -f bin/*

build-fetch-publication-pages: ## Build and deploy fetch-publication-pages
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/fetch-publication-pages ./cmd/fetch-publication-pages/

build-parse-publication-pages: ## Build and deploy parse-publication-pages
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/parse-publication-pages ./cmd/parse-publication-pages/

build-push-publications: ## Build and deploy push-publications
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/push-publications ./cmd/push-publications/