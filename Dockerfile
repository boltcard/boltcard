FROM golang:1.19.0-bullseye

WORKDIR /App
ADD . /App
RUN go build

WORKDIR /App/createboltcard
RUN go get github.com/skip2/go-qrcode
RUN go build

WORKDIR /App

ENTRYPOINT ["/App/boltcard"]