build:
	@go build -o api ./*.go

run: build
	@go run *.go

clean: 
	@go clean