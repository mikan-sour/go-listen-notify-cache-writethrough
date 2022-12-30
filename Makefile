BINARY_NAME=ListenForDBEvent

down:
	docker-compose down -v
	rm -rf ./data
up:
	export DOCKER_CLIENT_TIMEOUT=120
	export COMPOSE_HTTP_TIMEOUT=120
	docker-compose up -d

build:
	@go build -o ${BINARY_NAME} src/main.go

run:
	@make up
	sleep 20
	@make build
	./${BINARY_NAME}

env:
	cp ./.env.sample ./.env

one:
	@sh scripts/test.sh 

test:
	@make down
	@make up
	sleep 30
	@go test -coverprofile=coverage.out ./src/...
	sleep 3
	@go tool cover -html=coverage.out 
	@make down
	rm coverage.out listener

