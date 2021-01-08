test:
	@go test ./service/...

componenttests:
	@go test ./component-tests/... -count=1
