
env: ## Create testing environment
	docker-compose up -d

dep: ## Ensure dependencies
	@go get -u golang.org/x/lint/golint
	@go get -u github.com/actgardner/gogen-avro
	@dep ensure -vendor-only

models: ## Create models
	@gogen-avro --package=model pkg/model schemas/publication.avsc schemas/company.avsc

lint:
	golint pkg/* cmd/*

test: ## Run tests
	/usr/local/go/bin/go test ./...

clean: ## Clean project files
	rm -rfd vendor
	rm -f bin/*

build-fetch-publication-pages: ## Build and deploy fetch-publication-pages
	CGO_ENABLED=0 GOOS=linux /usr/local/go/bin/go build -a -installsuffix cgo -o ./bin/fetch-publication-pages ./cmd/fetch-publication-pages/

build-parse-publication-pages: ## Build and deploy parse-publication-pages
	CGO_ENABLED=0 GOOS=linux /usr/local/go/bin/go build -a -installsuffix cgo -o ./bin/parse-publication-pages ./cmd/parse-publication-pages/

build-push-publications: ## Build and deploy push-publications
	CGO_ENABLED=0 GOOS=linux /usr/local/go/bin/go build -a -installsuffix cgo -o ./bin/push-publications ./cmd/push-publications/