FROM busybox

FROM golang:1.16.2 as backend
COPY . /build
WORKDIR /build
RUN go mod download && \
    CGO_ENABLED=1 GOOS=linux go build -installsuffix cgo -ldflags="-s -w -linkmode external -extldflags -static" -o gokanban .

RUN git clone https://github.com/golang-migrate/migrate.git
WORKDIR /build/migrate
RUN CGO_ENABLED=1 go build -ldflags='-s -w -linkmode external -extldflags -static' -tags 'sqlite3 file' ./cmd/migrate/

FROM node:6 as frontend
COPY ./frontend /frontend
RUN (cd /frontend && npm i && PRODUCTION=true ./node_modules/.bin/gulp build)

FROM scratch
COPY --from=backend /build/gokanban /gokanban
COPY --from=backend /build/migrations /migrations
COPY --from=backend /build/migrate/migrate /migrate
COPY --from=frontend /static /static
COPY --from=frontend /frontend/templates/index.html /frontend/templates/index.html
COPY --from=frontend /frontend/templates/login.html /frontend/templates/login.html
COPY --from=busybox /bin/busybox /bin/busybox
COPY --from=busybox /bin/sh /bin/sh

EXPOSE 8080

CMD /migrate -path /migrations/ -database "sqlite3://${GOKANBAN_DB_FILE}" up; /gokanban
