GOPATH?=$(shell go env GOPATH)


build:
	go build -ldflags="-s -w" goctl.go

mac:
	GOOS=darwin go build -ldflags="-s -w" -o goctl-darwin goctl.go

win:
	GOOS=windows go build -ldflags="-s -w" -o goctl.exe goctl.go
	$(if $(shell command -v upx), upx goctl.exe)

linux:
	GOOS=linux go build -ldflags="-s -w" -o gogen goctl.go
	cp gogen $(GOPATH)/bin

image:
	docker build --rm --platform linux/amd64 -t kevinwan/goctl:$(version) .
	docker tag kevinwan/goctl:$(version) kevinwan/goctl:latest
	docker push kevinwan/goctl:$(version)
	docker push kevinwan/goctl:latest
	docker build --rm --platform linux/arm64 -t kevinwan/goctl:$(version)-arm64 .
	docker tag kevinwan/goctl:$(version)-arm64 kevinwan/goctl:latest-arm64
	docker push kevinwan/goctl:$(version)-arm64
	docker push kevinwan/goctl:latest-arm64
