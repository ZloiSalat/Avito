build:
	@go build -o GolangProjects/Avito

run: build
	@./GolangProjects/Avito

test:
	@go test -v ./..