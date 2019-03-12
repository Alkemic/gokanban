# goKanban

Simple Kanban project build with Go and AngularJS.

## Build

Checkout code, out side `$GOPATH`, as this project uses go modules, and simply build.

```sh
~ $ git clone https://github.com/Alkemic/gokanban.git
~ $ cd gokanban
~/gokanban $ go build .
```

And install frontend requirements.

```sh
~/gokanban $ cd frontend
~/gokanban/frontend $ npm i
```

##

## Migrations

To migrate database it's require to build golang-migrate with support of SQLite, which is not supported by their official releases.

First checkout latest code **out side** of `$GOPATH`, then navigate to `cmd/migrate` and build with at least `sqlite3` tag (in following example built with SQLite, PostgreSQL and MySQL support).

```sh
~ $ git clone https://github.com/golang-migrate/migrate.git
~ $ cd migrate/cmd/migrate
~/migrate/cmd/migrate $ go build -tags 'mysql sqlite3 postgres' -ldflags="-X main.Version=$(git describe --tags)" -o $GOPATH/bin/migrate .
````

### Migrate

To migrate just run following command within source root directory.

```
migrate -path ./migrations -database sqlite3://db.sqlite up
```
