# syntax=docker/dockerfile:1.2.1-labs@sha256:600e5c62eedff338b3f7a0850beb7c05866e0ef27b2d2e8c02aa468e78496ff5
FROM --platform=linux/amd64 golang:1.26-alpine@sha256:3082400e369fa24d5fc60bca20edab3f6d604e0c5a690ec66b295eff4dd87ade AS build

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

FROM alpine:3.23@sha256:85fe1e81d6758c208f3e1eed4338a1997e19d4be002d4dd32d3100c9a8c010a0 AS runtime

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
