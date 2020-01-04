.DEFAULT_GOAL := build

build: packages build-fetch-company-pages build-fetch-publication-pages build-parse-publications ## Build all binaries

packages: ## Download packages
	@go mod download

models: ## Create models
	@gogen-avro --package=model pkg/model \
		schemas/publication.avsc \
		schemas/page.avsc \
		schemas/company.avsc

dep: ## Download dependencies
	@go get -d github.com/actgardner/gogen-avro/...
	@go install github.com/actgardner/gogen-avro/gogen-avro

lint:
	@golint pkg/* cmd/*

test: ## Run tests
	@go test ./...

clean: ## Clean project files
	@rm -f bin/*

build-fetch-publication-pages: ## Build and deploy fetch-publication-pages
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/fetch-publication-pages ./cmd/fetch-publication-pages/

build-parse-publications: ## Build and deploy parse-publications
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/parse-publications ./cmd/parse-publications/

build-fetch-company-pages: ## Build and deploy fetch-publication-pages
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/fetch-company-pages ./cmd/fetch-company-pages/