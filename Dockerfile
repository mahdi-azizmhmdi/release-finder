FROM harbor-dev.okala.com/docker.io/golang:1.23.1 AS builder

ARG MAIN_APP_PATH=cmd/main.go

WORKDIR /app


COPY . .

RUN go install -mod=vendor -v -ldflags "-s" ./cmd/...
RUN go build -o main ${MAIN_APP_PATH}


FROM harbor-dev.okala.com/docker.io/alpine:latest

WORKDIR /app

COPY --from=builder /app/main /app
# COPY ./configs.yml /app/  this has added in helm chart as configmap

RUN apk update && \
    apk add --no-cache ca-certificates gcompat && \
    update-ca-certificates

ENTRYPOINT ["/app/main"]