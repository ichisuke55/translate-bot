FROM golang:1.21.9-alpine3.19 AS go
WORKDIR /app
COPY go.mod go.sum main.go ./
COPY config/ ./config
RUN go mod download \
&& go build -o main /app/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=go /app/main .
USER 1001
CMD ["/app/main"]
