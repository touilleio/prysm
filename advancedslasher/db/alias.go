package db

import "github.com/prysmaticlabs/prysm/advancedslasher/db/iface"

// Database defines the necessary methods for the Slasher's DB which may be implemented by any
// key-value or relational database in practice. This is the full database interface which should
// not be used often. Prefer a more restrictive interface in this package.
type Database = iface.Database
