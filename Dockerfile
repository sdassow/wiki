FROM golang:alpine
RUN apk update && apk add --update git
RUN go get -v github.com/GeertJohan/go.rice/rice
RUN go get -v github.com/openmicroapps/wiki
EXPOSE 8000
RUN mkdir /data
CMD ["/go/bin/wiki", "--data", "/data"]
