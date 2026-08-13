FROM golang:1.24.12 as build

ARG BRANCH
ARG RELEASE_TAG
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=${TARGETOS} \
  GOARCH=${TARGETARCH} \
  GOARM=${TARGETVARIANT} \
  DEBIAN_FRONTEND=noninteractive \
  PATH="/root/go/bin:${PATH}" \
  BRANCH=${BRANCH} \
  RELEASE_TAG=${RELEASE_TAG}

WORKDIR /go/src/github.com/openebs/dynamic-localpv-provisioner/

RUN apt-get update && apt-get install -y make git

COPY go.mod go.sum ./
# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

COPY . .

RUN make provisioner-localpv

FROM alpine:3.23.5

ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
LABEL org.label-schema.name="provisioner-localpv"
LABEL org.label-schema.description="Dynamic Local PV Provisioner for OpenEBS"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

RUN apk add --no-cache \
    iproute2 \
    bash \
    curl \
    net-tools \
    procps \
    ca-certificates \
    util-linux \
    e2fsprogs \
    xfsprogs \
    xfsprogs-extra \
    blkid \
    findmnt \
    quota-tools

COPY --from=build /go/src/github.com/openebs/dynamic-localpv-provisioner/bin/provisioner-localpv/provisioner-localpv /usr/local/bin/provisioner-localpv

ENTRYPOINT ["/usr/local/bin/provisioner-localpv"]
