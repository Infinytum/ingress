FROM gcr.io/distroless/static-debian11:nonroot
ARG TARGETPLATFORM

EXPOSE 8080 8443
COPY bin/${TARGETPLATFORM}/ingress /usr/bin/ingress

ENTRYPOINT [ "/usr/bin/ingress" ]