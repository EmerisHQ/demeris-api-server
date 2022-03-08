FROM golang:1.17 as builder

ARG GIT_TOKEN

RUN go env -w GOPRIVATE=github.com/emerishq/*
RUN git config --global url."https://git:${GIT_TOKEN}@github.com".insteadOf "https://github.com"

WORKDIR /app
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 GOPROXY=direct make

FROM alpine:latest

RUN apk --no-cache add ca-certificates mailcap && addgroup -S app && adduser -S app -G app
COPY --from=builder /app/build/api-server /usr/local/bin/api-server
USER app
ENTRYPOINT ["/usr/local/bin/api-server"]