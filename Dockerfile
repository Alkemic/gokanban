FROM golang:1.10
COPY . /go/src/github.com/Alkemic/gokanban/
RUN go get -u github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/Alkemic/gokanban/
RUN dep ensure && go test -cover ./... && go install ./...
RUN strip /go/bin/gokanban

FROM node:8
COPY ./frontend /frontend
RUN (cd /frontend && npm i)

FROM debian:9
COPY --from=0 /go/bin/gokanban /gokanban
COPY --from=1 /frontend /frontend

EXPOSE 8080

CMD "/gokanban"
