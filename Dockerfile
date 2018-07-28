FROM golang:1.10
RUN go get -v github.com/Alkemic/gokanban
RUN strip /go/bin/gokanban

FROM node:8
COPY --from=0 /go/src/github.com/Alkemic/gokanban/frontend /frontend
RUN (cd /frontend && npm i)

FROM debian:9
COPY --from=0 /go/bin/gokanban /gokanban
COPY --from=1 /frontend /frontend

EXPOSE 8080

CMD "/gokanban"
