OBJS = $(shell find cmd -mindepth 1 -type d -execdir printf '%s\n' {} +)
OBJS_CI = $(shell find cmd -mindepth 1 -type d -execdir printf '%s-ci\n' {} +)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
BASEPKG = github.com/allinbits/demeris-api-server
EXTRAFLAGS :=

.PHONY: $(OBJS) $(OBJS_CI) lint clean generate-swagger

all: $(OBJS)

clean:
	@rm -rf build docs/swagger.* docs/docs.go

generate-swagger:
	go generate ${BASEPKG}/docs
	@rm docs/docs.go

$(OBJS):
	go build -o build/$@ -ldflags='-X main.Version=${BRANCH}-${COMMIT}' ${EXTRAFLAGS} ${BASEPKG}/cmd/$@

$(OBJS_CI): lint
	TARGET=$$(echo $@ | sed 's|-ci||g'); \
	go build -o build/$${TARGET} -ldflags='-X main.Version=${BRANCH}-${COMMIT}' ${EXTRAFLAGS} ${BASEPKG}/cmd/$${TARGET}

lint:
	golangci-lint run ./...
