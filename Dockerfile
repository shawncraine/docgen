# BUILD image in golang:alpine
FROM golang:alpine
RUN apk add --no-cache --update git
ENV GOPATH=/go

RUN go get -u github.com/golang/dep/cmd/dep && \
    go get -u github.com/jteeuwen/go-bindata/...
COPY . $GOPATH/src/github.com/thedevsaddam/docgen
WORKDIR $GOPATH/src/github.com/thedevsaddam/docgen

ENV GOARCH="amd64"
ENV GOOS="linux"
ENV CGO_ENABLED=0

RUN dep ensure -v -vendor-only && \
    go-bindata assets/ && \
    go install -v

# Copy the binary and the config to new image
FROM alpine:latest 
RUN apk add vim && \
    apk add --no-cache --update ca-certificates openssl && \
    apk add --no-cache tzdata
LABEL maintainer "thedevsaddam@gmail.com"
COPY --from=0 /go/bin/docgen /usr/local/bin/docgen
ADD ./files/config.yaml /files/config.yaml
EXPOSE 9000
ENTRYPOINT ["docgen"]
