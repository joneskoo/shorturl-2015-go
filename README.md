yx.fi shorturl - go iteration
=============================

Yet another rewrite of yx.fi shorturl, this time in go.

Structure:

yxfi_backend/
    Runnable HTTP API backend service. Listens for HTTP
    requests and implements yx.fi API.

shorturl/
    Database model and representation of short URLs
