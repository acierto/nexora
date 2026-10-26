# Order listing API

The order listing endpoint uses stable cursor pagination.

## Cursor ordering and filters

Sort by created_at descending and order_id descending. The cursor carries the
last tuple, a format version, and a filter fingerprint. Validate all fields and
reject a cursor reused with different filters. Tenant identity always comes from
the verified caller. Fetch limit plus one rows to decide whether next_cursor exists.

## Equal timestamp regression

Seed three orders with identical timestamps and request pages of size two. The
second page must contain the remaining order without repeating either earlier ID.
Also test a malformed cursor, a changed filter, and a new order inserted between
requests. The page size default is 25 and the accepted range is 1 through 100.
