yx.fi shorturl - go iteration
=============================

Yet another rewrite of yx.fi shorturl, this time in go.

DB Schema:
```sql
CREATE TABLE shorturl (
    id SERIAL PRIMARY KEY,
    url text,
    ts timestamp without time zone DEFAULT now() NOT NULL,
    host text,
    cookie text
);
```

Note: we assume server is used behind reverse proxy. Ensure that the frontend
sets header X-Forwarded-Proto = https or http accordingly.
