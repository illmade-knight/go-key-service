format:
	@go fmt ./...
.PHONY: format

build:
	@go build ./...
.PHONY : build

test:
	@go test ./...

testv:
	@go test -v -cover -count=1 ./...

start-service-simple:
	@go run cmd/simple/runkeystore.go

start-service-scalable:
	@go run cmd/keyservice/runscalablekeyservice.go	
