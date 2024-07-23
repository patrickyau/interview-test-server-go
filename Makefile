all: install generate run

install:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

generate:
	oapi-codegen -config server.cfg.yaml ./api/openAPI.yaml

run:
	go run app/main.go

visit:
	open http://localhost:8000/swagger/
