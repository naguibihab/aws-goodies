build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/calc calc/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/logs logs/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/orders orders/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/events events/main.go

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock

.PHONY: deploy
deploy: clean build
	sls deploy --verbose
