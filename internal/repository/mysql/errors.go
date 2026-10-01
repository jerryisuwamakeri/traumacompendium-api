package mysql

import "errors"

// ErrDuplicateEntry wraps any MySQL ER_DUP_ENTRY (1062) violation of a
// unique constraint. Service layers check for it with errors.Is to return
// a 409 Conflict instead of a generic 500.
var ErrDuplicateEntry = errors.New("mysql: duplicate entry")
