FROM golang:1.17-alpine

# Install make
RUN apk update && apk add make && apk add build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV E2E=1

CMD go test -coverprofile=e2e-coverage.out ./src/... && go tool cover -html e2e-coverage.out -o e2e-coverage.html

# ENTRYPOINT ["tail", "-f", "/dev/null"]