.PHONY: test lint mocks

test:
	go test --race ./...

lint:
	golangci-lint run

mocks:
	mockery

deploy:
	git pull && docker compose pull && docker stack deploy -c docker-compose.yml app
