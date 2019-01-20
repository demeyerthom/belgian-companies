
env: ## Create testing environment
	COMPOSE_PROJECT_NAME=testing docker-compose up -d

dep: ## Ensure dependencies
	@go get -u golang.org/x/lint/golint
	@go get -u github.com/actgardner/gogen-avro
	@dep ensure -vendor-only

models: ## Create models
	@gogen-avro --package=model pkg/model schemas/publication.avsc schemas/publication-page.avsc schemas/company.avsc

lint:
	golint pkg/* cmd/*

test: ## Run tests
	/usr/local/go/bin/go test ./...

clean: ## Clean project files
	rm -rfd vendor
	rm -f bin/*

build-fetch-publication-pages: clean dep models ## Build and deploy fetch-publication-pages
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/fetch-publication-pages ./cmd/fetch-publication-pages/

build-parse-publications: clean dep models ## Build and deploy parse-publications
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/parse-publications ./cmd/parse-publications/

build-project-publications: clean dep models ## Build and deploy project-publications
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/project-publications ./cmd/project-publications/