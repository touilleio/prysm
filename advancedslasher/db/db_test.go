package db

import "github.com/prysmaticlabs/prysm/advancedslasher/db/kv"

var _ Database = (*kv.Store)(nil)
