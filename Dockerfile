FROM golang:1.18 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/app

FROM --platform=linux/amd64 gcr.io/distroless/static-debian11

COPY --from=build /go/bin/app /
EXPOSE 80 443
ENTRYPOINT [ "/app" ]
