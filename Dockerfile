FROM gcr.io/distroless/static-debian11
ARG TARGETPLATFORM

EXPOSE 80 443
COPY bin/${TARGETPLATFORM}/ingress /

ENTRYPOINT [ "/ingress" ]