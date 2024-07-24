all: install generate run

install:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

generate:
	oapi-codegen -config server.cfg.yaml ./api/openAPI.yaml

# build:
# 	go build -o /interview-test-server ./...
#   docker build --tag interview-test-server-go .

run:
	# docker run --rm -p 8080:8080 --name interview-test-server-go interview-test-server-go
	go run ./...

visit:
	open http://localhost:8000/swagger/
