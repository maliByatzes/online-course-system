CREATE TABLE students (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  name       TEXT NOT NULL,
  email      TEXT NOT NULL UNIQUE,
  api_key    TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE auths (
  id               INTEGER PRIMARY KEY AUTOINCREMENT,
  student_id       INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  source           TEXT NOT NULL,
  source_id        TEXT NOT NULL,
  access_token     TEXT NOT NULL,
  refresh_token    TEXT NOT NULL,
  expiry           TEXT,
  created_at       TEXT NOT NULL,
  updated_at       TEXT NOT NULL,

  UNIQUE(student_id, source),
  UNIQUE(source, source_id)
);
