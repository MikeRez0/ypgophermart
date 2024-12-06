GOLANGCI_LINT_CACHE?=praktikum-golangci-lint-cache
current_dir := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))

.PHONY: golangci-lint-run
golangci-lint-run:
	-docker run --rm -v .:/source -v $(GOLANGCI_LINT_CACHE):/root/.cache -w //source golangci/golangci-lint golangci-lint run -c .golangci.yml
	-bash -c 'cat ./.golangci-lint/report-unformatted.json | jq > ./.golangci-lint/report.json'


.PHONY: build
build:
	go build -C cmd/gophermart

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


.PHONY: yptestdbstart
yptestdbstart:
	docker run --rm \
		--name=praktikum-db \
		--expose 5433 \
		-e POSTGRES_PASSWORD="postgres" \
		-e POSTGRES_USER="postgres" \
		-e POSTGRES_DB="praktikum" \
		-d \
		-p 5433:5432 \
		postgres:15.3

.PHONY: yptest
yptest: build
	./.yptest/gophermarttest \
		-test.v -test.run=^TestGophermart$ \
		-gophermart-binary-path=cmd/gophermart/gophermart \
		-gophermart-host=localhost \
		-gophermart-port=9991 \
		-gophermart-database-uri="postgresql://postgres:postgres@localhost:5433/praktikum?sslmode=disable" \
		-accrual-binary-path=cmd/accrual/accrual_darwin_amd64 \
		-accrual-host=localhost \
		-accrual-port=9990 \
		-accrual-database-uri="postgresql://postgres:postgres@localhost:5433/praktikum?sslmode=disable" > ./.yptest/log.txt