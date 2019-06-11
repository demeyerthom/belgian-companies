.PHONY: vendor dependencies tidy env models lint test clean

vendor: go.sum
	@go mod download
	@go mod vendor

tidy:
	@go mod tidy

dependencies:
	@go get -u github.com/actgardner/gogen-avro/gogen-avro

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
