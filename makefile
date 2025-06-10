local-build:
	docker build -t mobydeck/ci-lark-notification .
	docker image prune -f

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.txt

coverhtml: coverage
	go tool cover -html=coverage.txt
