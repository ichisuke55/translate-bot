FROM golang:1.21.9-alpine3.19 AS go
WORKDIR /app
COPY go.mod go.sum main.go ./
COPY config/ ./config
COPY logging/ ./logging
RUN go mod download \
    && go build -o main /app/main.go

FROM alpine:3.19
WORKDIR /app
RUN apk --update add tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    apk del tzdata && \
    rm -rf /var/cache/apk/* && \
    mkdir log && \
    chown 1001 -R /app
COPY --from=go /app/main .
USER 1001
CMD ["/app/main"]
