FROM busybox

FROM golang:1.15 as backend
COPY . /build
WORKDIR /build
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o gokanban . && \
    strip gokanban
ADD https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz migrate.linux-amd64.tar.gz
RUN tar zxvf migrate.linux-amd64.tar.gz

FROM node:6 as frontend
COPY ./frontend /frontend
RUN (cd /frontend && npm i && PRODUCTION=true ./node_modules/.bin/gulp build)

FROM scratch
COPY --from=backend /build/gokanban /gokanban
COPY --from=frontend /static /static
COPY --from=frontend /frontend/templates/index.html /frontend/templates/index.html
COPY --from=busybox /bin/busybox /bin/busybox
COPY --from=busybox /bin/sh /bin/sh

EXPOSE 8080

CMD "/gokanban"
CMD /migrate -path /migrations/ -database "sqlite://${GOKANBAN_DB_FILE}" up; /webrss
