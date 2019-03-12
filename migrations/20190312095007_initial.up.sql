PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS "column" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "name" VARCHAR(127),
  "limit" INTEGER DEFAULT 10,
  "position" INTEGER DEFAULT 0,
  "created_at" datetime,
  "updated_at" datetime,
  "deleted_at" datetime NULL
);
CREATE INDEX idx_column_deleted_at ON "column"("deleted_at") ;

CREATE TABLE IF NOT EXISTS "task" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "title" VARCHAR(255),
  "description" VARCHAR(255) ,
  "column_id" INTEGER,
  "position" INTEGER DEFAULT 0,
  "color" VARCHAR(7),
  "created_at" DATETIME,
  "updated_at" DATETIME,
  "deleted_at" DATETIME NULL,
  FOREIGN KEY("column_id") REFERENCES "column"("id")
);
CREATE INDEX idx_task_deleted_at ON "task"("deleted_at") ;

CREATE TABLE IF NOT EXISTS "tag" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "name" VARCHAR(127)
);

CREATE TABLE IF NOT EXISTS "task_tags" (
  "task_id" INTEGER NOT NULL,
  "tag_id" INTEGER NOT NULL,
  FOREIGN KEY("task_id") REFERENCES "task"("id"),
  FOREIGN KEY("tag_id") REFERENCES "tag"("id"),
  PRIMARY KEY ("task_id","tag_id")
);

CREATE TABLE IF NOT EXISTS "task_log" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "task_id" INTEGER NOT NULL,
  "old_column_id" INTEGER NOT NULL,
  "action" VARCHAR(255),
  "created_at" DATETIME,
  FOREIGN KEY("old_column_id") REFERENCES "column"("id"),
  FOREIGN KEY("task_id") REFERENCES "task"("id")
);
