fmt:
	command -v gofumpt || (WORK=$(shell pwd) && cd /tmp && GO111MODULE=on go get mvdan.cc/gofumpt && cd $(WORK))
	gofumpt -w -s -d .

test:
	go test -v ./tests/...

lint:
	golangci-lint run  -v

ci/lint: export GO111MODULE=on
ci/lint: export GOPROXY=https://goproxy.io,direct
ci/lint: export GOOS=linux
ci/lint: export CGO_ENABLED=0
ci/lint: lint