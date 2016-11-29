FROM golang:1.7.3-alpine

ADD . /go/src/github.com/zkmoney/topthree
WORKDIR /go/src/github.com/zkmoney/topthree

RUN go build -o /build/topthree

EXPOSE 8080

CMD ["/build/topthree"]
