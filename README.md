# goKanban

Simple Kanban project build with Go and AngularJS.

## Build

Checkout code, outside `$GOPATH`.

```sh
~ $ git clone https://github.com/Alkemic/gokanban.git
~ $ cd gokanban
~/gokanban $ go build .
```

And install frontend requirements.

```sh
~/gokanban $ cd frontend
~/gokanban/frontend $ npm i
~/gokanban/frontend $ ./node_modules/.bin/gulp build
```

## Migrate database

* Install [golang migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation)
* Migrate using ``migrate -path ./migrations -database sqlite3://db.sqlite up``

## Auth

* Only single user
* Default email / password is `admin` / `admin`
