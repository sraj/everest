package model

import "github.com/lib/pq"

// Tags is a PostgreSQL text array scanned via pq.Array.
type Tags []string

func (t *Tags) Scan(src any) error   { return (*pq.StringArray)((*[]string)(t)).Scan(src) }
