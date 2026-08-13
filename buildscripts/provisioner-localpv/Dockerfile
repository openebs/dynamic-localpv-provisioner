FROM alpine:3.23.5

RUN apk add --no-cache \
    iproute2 \
    bash \
    curl \
    net-tools \
    mii-tool \
    procps \
    libc6-compat \
    ca-certificates \
    util-linux \
    e2fsprogs \
    xfsprogs \
    xfsprogs-extra \
    blkid \
    findmnt \
    quota-tools

COPY provisioner-localpv /

ARG ARCH
ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
LABEL org.label-schema.name="provisioner-localpv"
LABEL org.label-schema.description="Dynamic Local PV Provisioner for OpenEBS"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

ENTRYPOINT ["/provisioner-localpv"]
