# Infinytum Ingress

A fast [caddy-based](https://github.com/caddyserver/caddy) ingress controller for Kubernetes

## How to build

To build this image for docker you need to specify the GOOS and output directory: 

```console
$ GOOS=linux go build -o bin/linux/arm64/ingress
$ docker build -t infinytum/ingress:latest .
```