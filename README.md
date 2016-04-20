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

TODO: user-unique cookie generation + store to db