
build-publication-pages:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/fetch-publication-pages ./cmd/fetch-publication-pages/
