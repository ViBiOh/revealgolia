FROM vibioh/scratch

ENTRYPOINT [ "/revealgolia" ]

ARG VERSION
ENV VERSION=${VERSION}

ARG TARGETOS
ARG TARGETARCH

COPY cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY release/revealgolia_${TARGETOS}_${TARGETARCH} /revealgolia
