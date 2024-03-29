OBJS = $(shell find cmd -mindepth 1 -type d -execdir printf '%s\n' {} +)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
BASEPKG = github.com/emerishq/demeris-api-server
EXTRAFLAGS :=

.PHONY: $(OBJS) clean generate-swagger

all: $(OBJS)

clean:
	@rm -rf build docs/swagger.* docs/docs.go

generate-swagger:
	go generate ${BASEPKG}/docs
	@rm docs/docs.go

test:
	go test -v -race ./... -cover

lint:
	golangci-lint run ./...

generate-mocks:
	go install github.com/golang/mock/mockgen
	go generate ./api/...
	go generate ./usecase/
	go install github.com/vektra/mockery/v2
	-@rm mocks/*.go
	mockery --srcpkg sigs.k8s.io/controller-runtime/pkg/client --name Client
	mockery --srcpkg k8s.io/client-go/informers --name GenericInformer
	mockery -r --dir lib --all --with-expecter

$(OBJS):
	go build -o build/$@ -ldflags='-X main.Version=${BRANCH}-${COMMIT}' ${EXTRAFLAGS} ${BASEPKG}/cmd/$@
