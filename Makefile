#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

app_container_name := app
docker_bin := $(shell command -v docker 2> /dev/null)
docker_compose_bin := $(shell command -v docker-compose 2> /dev/null)
docker_compose_yml := docker/docker-compose.yml
user_id := $(shell id -u)

.PHONY : help pull build push login test clean \
         app-pull app app-push\
         sources-pull sources sources-push\
         nginx-pull nginx nginx-push\
         up down restart shell install
.DEFAULT_GOAL := help

# --- [ Development tasks ] -------------------------------------------------------------------------------------------
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build containers
	$(docker_compose_bin) --file "$(docker_compose_yml)" build

up: build ## Run app
	$(docker_compose_bin) --file "$(docker_compose_yml)" up

doc: build ## Run local documentation
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm -p 8090:8090 $(app_container_name) godoc -http=:8090 -goroot="/usr"
	## "http://localhost:8090/pkg/?m=all"

mock: ## Generate mocks
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) mockgen -destination=internal/app/mocks/repository.go -package=mocks github.com/belamov/ypgo-url-shortener/internal/app/storage Repository
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) mockgen -destination=internal/app/mocks/generator.go -package=mocks github.com/belamov/ypgo-url-shortener/internal/app/services/generator URLGenerator
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) mockgen -destination=internal/app/mocks/random.go -package=mocks github.com/belamov/ypgo-url-shortener/internal/app/services/random Generator
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) mockgen -destination=internal/app/mocks/ipchecker.go -package=mocks github.com/belamov/ypgo-url-shortener/internal/app/services IPCheckerInterface
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) mockgen -destination=internal/app/mocks/shortener.go -package=mocks github.com/belamov/ypgo-url-shortener/internal/app/services ShortenerInterface
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) mockgen -destination=internal/app/mocks/cryptographer.go -package=mocks github.com/belamov/ypgo-url-shortener/internal/app/services/crypto Cryptographer

proto: ## Generate proto files
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) protoc --go_out=. --go_opt=paths=source_relative \
                                                                                          --go-grpc_out=. --go-grpc_opt=paths=source_relative \
                                                                                          internal/app/proto/*.proto

lint:
	$(docker_bin) run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run

fieldaligment-fix:
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) fieldalignment -fix ./... || true

gofumpt:
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) gofumpt -l -w .

test: ## Execute tests
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) go test -v -race ./...

check: build proto fieldaligment-fix gofumpt lint test  ## Run tests and code analysis

staticlint:
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) go build -v -o /usr/src/app/cmd/staticlint/staticlint /usr/src/app/cmd/staticlint
	$(docker_compose_bin) --file "$(docker_compose_yml)" run --rm $(app_container_name) /usr/src/app/cmd/staticlint/staticlint ./...

# Prompt to continue
prompt-continue:
	@while [ -z "$$CONTINUE" ]; do \
		read -r -p "Would you like to continue? [y]" CONTINUE; \
	done ; \
	if [ ! $$CONTINUE == "y" ]; then \
        echo "Exiting." ; \
        exit 1 ; \
    fi
