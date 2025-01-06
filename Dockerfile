FROM golang:1.23-alpine AS builder

WORKDIR /app

ADD https://obs-community-intl.obs.ap-southeast-1.myhuaweicloud.com/obsutil/current/obsutil_linux_amd64.tar.gz ./
RUN tar xzf ./obsutil_linux_amd64.tar.gz

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-w -s" -o ./temporal-shell ./main.go

FROM alpine:3.21

ENV TZ=Asia/Shanghai \
    REGISTRY_AUTH_FILE=/tmp/auth.json

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone && \
    apk add --no-cache  \
        bash \
        skopeo

COPY --from=builder /app/temporal-shell /app/obsutil_linux_amd64_*/obsutil /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/temporal-shell"]