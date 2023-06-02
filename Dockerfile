FROM cgr.dev/chainguard/static
ARG TARGETPLATFORM

EXPOSE 8080 8443
COPY bin/${TARGETPLATFORM}/ingress /usr/bin/ingress

ENTRYPOINT [ "/usr/bin/ingress" ]