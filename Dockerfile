# syntax=docker/dockerfile:1.29.0-labs@sha256:36366b12c1dffb228bb25e9feffb73707b93388f084f4927d2b7686dba29e76a
FROM --platform=linux/amd64 golang:1.27.4-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS build

ARG PKG_NAME
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH

RUN apk add --no-cache make git

WORKDIR /src
COPY go.mod ./
COPY go.sum ./

RUN --mount=type=cache,id=go-build-${TARGETOS}-${TARGETARCH}${TARGETVARIANT},target=/root/.cache/go-build \
	--mount=type=cache,id=go-pkg-${TARGETOS}-${TARGETARCH}${TARGETVARIANT},target=/go/pkg \
		go mod download -x

COPY . ./

RUN --mount=type=cache,id=go-build-${TARGETOS}-${TARGETARCH}${TARGETVARIANT},target=/root/.cache/go-build \
	--mount=type=cache,id=go-pkg-${TARGETOS}-${TARGETARCH}${TARGETVARIANT},target=/go/pkg \
		go build -ldflags "-w -s" \
			-o bin/${PKG_NAME} \
			./cmd/${PKG_NAME}
RUN mv bin/${PKG_NAME}* /bin/

FROM alpine:3.24@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf AS runtime

ARG PKG_NAME
ARG VCS_REF
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

LABEL org.opencontainers.image.revision=$VCS_REF \
	org.opencontainers.image.source="https://github.com/hairyhenderson/${PKG_NAME}"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/${PKG_NAME} /${PKG_NAME}

ENTRYPOINT [ "/jarvis_exporter" ]
