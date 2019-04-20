
env: ## Create testing environment
	docker-compose up -d

models: ## Create models
	@gogen-avro --package=model pkg/model \
		schemas/publication.avsc \
		schemas/publication-page.avsc \
		schemas/company.avsc \
		schemas/address.avsc \
		schemas/company_page.avsc \
		schemas/company_pages.avsc \
		schemas/financial_information.avsc \
		schemas/legal_function.avsc \
		schemas/nace_code.avsc

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

build-project-publications: ## Build and deploy project-publications
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/project-publications ./cmd/project-publications/

build-fetch-company-pages: ## Build and deploy fetch-publication-pages
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/fetch-company-pages ./cmd/fetch-company-pages/