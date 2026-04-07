FROM rg.fr-par.scw.cloud/vibioh/scratch

ENTRYPOINT [ "/revealgolia" ]

ARG VERSION
ENV VERSION=${VERSION}

ARG GIT_SHA
ENV GIT_SHA=${GIT_SHA}

ARG TARGETOS
ARG TARGETARCH

COPY cacert.pem /etc/ssl/cert.pem
COPY release/revealgolia_${TARGETOS}_${TARGETARCH} /revealgolia
