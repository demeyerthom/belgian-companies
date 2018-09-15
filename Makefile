
env: ## Create testing environment
	@docker-compose --file deployments/docker-compose.yaml up -d

dep: ## Ensure dependencies
	@go get github.com/golang/lint
	@dep ensure -vendor-only

test: dep ## Run tests
	@go test ./...

clean: ## Clean project files
	rm -rfd vendor
	rm -f bin/*

build-fetch-publication-pages: clean dep ## Build and deploy fetch-publication-pages
	@sh scripts/build-fetch-publication-pages.sh

build-parse-publication-pages: clean dep ## Build and deploy parse-publication-pages
	@sh scripts/build-parse-publication-pages.sh

build-push-publications: clean dep ## Build and deploy push-publications
	@sh scripts/build-push-publications.sh

build-all: clean dep build-fetch-publication-pages build-parse-publication-pages build-push-publications
