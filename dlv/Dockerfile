FROM golang:alpine

RUN go install github.com/go-delve/delve/cmd/dlv@latest

ENTRYPOINT [ "dlv" ]
