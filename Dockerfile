FROM golang:1.11
COPY . /gokanban/
WORKDIR /gokanban/
RUN go mod download && go test -cover ./... && go install ./...
RUN strip /go/bin/gokanban

FROM node:8
COPY ./frontend /frontend
RUN (cd /frontend && npm i)

FROM alpine:3.8
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY --from=0 /go/bin/gokanban /gokanban
COPY --from=1 /frontend /frontend

EXPOSE 8080

CMD "/gokanban"
