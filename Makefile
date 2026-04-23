APP=imans
PKG=./cmd/imans

.PHONY: build test fmt vet generate schema

build:
	go build -ldflags "-X github.com/imans-ai/imans-cli/internal/version.Version=$${VERSION:-dev} -X github.com/imans-ai/imans-cli/internal/version.Commit=$$(git rev-parse --short HEAD 2>/dev/null || printf unknown) -X github.com/imans-ai/imans-cli/internal/version.BuildDate=$$(date -u +%Y-%m-%dT%H:%M:%SZ) -X github.com/imans-ai/imans-cli/internal/version.SchemaVersion=$${SCHEMA_VERSION:-dev}" -o $(APP) $(PKG)

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal ./api

vet:
	go vet ./...

generate:
	go generate ./...

schema:
	./scripts/refresh-schema.sh
