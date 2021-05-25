.PHONY: test

test:
	@go test -coverprofile bin/coverage.out ./...
	@go tool cover -html .\bin\coverage.out