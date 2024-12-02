GOLANGCI_LINT_CACHE?=praktikum-golangci-lint-cache
current_dir := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))

.PHONY: golangci-lint-run
golangci-lint-run:
	-docker run --rm -v .:/source -v $(GOLANGCI_LINT_CACHE):/root/.cache -w //source golangci/golangci-lint golangci-lint run -c .golangci.yml
	-bash -c 'cat ./.golangci-lint/report-unformatted.json | jq > ./.golangci-lint/report.json'


.PHONY: build-server
build-server:
	go build -C cmd/server

.PHONY: build-agent
build-agent:
	go build -C cmd/agent

.PHONY: test
test:
	go test ./...
.PHONY: test-cover
test-cover: test
	go test ./... -cover

.PHONY: db-start
db-start:
	docker compose -f "scripts/db/docker-compose.yaml" up -d --build

.PHONY: db-stop
db-stop:
	docker compose -f "scripts/db/docker-compose.yaml" down

.PHONY: db-clean
db-clean:
	sudo rm -rf ./.var/data/

.PHONY: db-migration-new
db-migration-new:
	docker run --rm \
    -v $(realpath ./internal/storage/migrations):/migrations \
    migrate/migrate:v4.16.2 \
        create \
        -dir /migrations \
        -ext .sql \
        -seq -digits 5 \
        $(name)

.PHONY: db-migration-up
db-migration-up:
	docker run --rm \
    -v $(realpath ./internal/storage/migrations):/migrations \
	--network host \
    migrate/migrate:v4.16.2 \
        -path=/migrations \
        -database postgres://metrics:metrics@localhost:5432/metrics_db?sslmode=disable \
        up 

.PHONY: db-migration-down
db-migration-down:
	docker run --rm \
    -v $(realpath ./internal/storage/migrations):/migrations \
	--network host \
    migrate/migrate:v4.16.2 \
        -path=/migrations \
        -database postgres://metrics:metrics@localhost:5432/metrics_db?sslmode=disable \
        down 1

.PHONY: swagger-editor
swagger-editor:
	docker run --rm \
	-d -p 80:8080 \
	-v $(realpath ./docs/swagger):/specs \
	-e SWAGGER_FILE=/specs/swagger.yaml \
	swaggerapi/swagger-editor

.PHONY: accrual-start
accrual-start:
	./cmd/accrual/accrual_darwin_amd64 -d "postgresql://gophermart:gophermart@localhost:5432/gophermart_db?sslmode=disable" -a localhost:8889