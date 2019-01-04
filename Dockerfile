FROM golang:1.11
COPY . /gokanban/
WORKDIR /gokanban/
RUN go mod download && go test -cover ./... && go install ./...
# RUN go test -cover ./... && go install ./...
RUN strip /go/bin/gokanban

FROM node:8
COPY ./frontend /frontend
RUN (cd /frontend && npm i)

FROM debian:9
COPY --from=0 /go/bin/gokanban /gokanban
COPY --from=1 /frontend /frontend

EXPOSE 8080

CMD "/gokanban"
