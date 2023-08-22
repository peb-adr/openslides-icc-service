FROM golang:1.20.7-alpine as base
WORKDIR /root/

RUN apk add git

COPY go.mod go.sum ./
RUN go mod download

COPY main.go main.go
COPY internal internal

# Build service in seperate stage.
FROM base as builder
RUN go build


# Development build.
FROM base as development

RUN ["go", "install", "github.com/githubnemo/CompileDaemon@latest"]
EXPOSE 9012

CMD CompileDaemon -log-prefix=false -build="go build" -command="./openslides-icc-service"


# Productive build
FROM scratch

LABEL org.opencontainers.image.title="OpenSlides ICC Service"
LABEL org.opencontainers.image.description="With the OpenSlides ICC Service clients can communicate with each other."
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/OpenSlides/openslides-icc-service"

COPY --from=builder /root/openslides-icc-service .
EXPOSE 9007
ENTRYPOINT ["/openslides-icc-service"]
HEALTHCHECK CMD ["/openslides-icc-service", "health"]
