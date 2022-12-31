BINARY_NAME=ListenForDBEvent

down:
	@docker-compose down -v
	rm -rf ./data
up:
	export DOCKER_CLIENT_TIMEOUT=120
	export COMPOSE_HTTP_TIMEOUT=120
	@docker-compose up -d postgres redis

build:
	@go build -o ${BINARY_NAME} src/main.go

run:
	@make build
	./${BINARY_NAME}

env:
	cp ./.env.sample ./.env

one:
	@sh scripts/test.sh 

test:
	@go test -coverprofile=coverage.out ./src/...
	@go tool cover -html=coverage.out 

e2e:
	@make down
	@docker-compose build
	@docker-compose up -d postgres redis
	sleep 30

	@docker-compose up blackbox

	


