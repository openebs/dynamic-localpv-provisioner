FROM alpine:3.18

RUN apk add --no-cache \
  iproute2 \
  bash \
  util-linux \
  e2fsprogs \
  xfsprogs \
  blkid \
  findmnt \
  ca-certificates

COPY pvc-manager /usr/local/bin/

ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
LABEL org.label-schema.name="pvc-manager"
LABEL org.label-schema.description="OpenEBS LocalPV PVC Manager"
LABEL org.label-schema.url="$DBUILD_SITE_URL"
LABEL org.label-schema.vcs-url="$DBUILD_REPO_URL"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date="$DBUILD_DATE"

ENTRYPOINT ["/usr/local/bin/pvc-manager"]
CMD ["--help"]