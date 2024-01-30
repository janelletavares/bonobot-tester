DOCKER_REPO ?= localhost:5001

api_run:
	go run main.go serve

api_build:
	CGO_ENABLED=0 go build -o api

docker_build:
	docker build . -t $(DOCKER_REPO)/api:latest

kind_push_image:
	kind load docker-image $(DOCKER_REPO)/api:latest --name zoom-participant-factory

create_migration:
	go run main.go migrate create $(name)

run_migrations:
	go run main.go migrate up

gofumpt:
	go install mvdan.cc/gofumpt@latest

fmt: gofumpt
	gofumpt -l -w .

lint:
	docker pull golangci/golangci-lint:latest
	docker run --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:latest golangci-lint run --timeout 5m -v

test:
	go test -v ./...