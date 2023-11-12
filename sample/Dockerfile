ARG BUILDER_IMAGE=golang:alpine
ARG RUNTIME_IMAGE=gcr.io/distroless/static


FROM $BUILDER_IMAGE as builder

        WORKDIR /app
        COPY . .

        RUN set -x && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
                go build -a -v \
			-trimpath \
			-tags osusergo,netgo,static_build \
			-o sample \
			.


FROM $RUNTIME_IMAGE

        WORKDIR /
        COPY --chown=nonroot:nonroot --from=builder /app/sample .
        USER nonroot:nonroot

        ENTRYPOINT ["/sample"]
