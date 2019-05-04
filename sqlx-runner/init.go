package runner

import (
	"database/sql"
	"time"

	"github.com/casualjim/dat"
	"github.com/casualjim/dat/kvs"
	"github.com/casualjim/dat/postgres"
	"github.com/cenkalti/backoff"
	"go.uber.org/zap"
)

var logger *zap.Logger

// LogQueriesThreshold is the threshold for logging "slow" queries
var LogQueriesThreshold time.Duration

// LogErrNoRows tells runner to log `sql.ErrNoRows`
var LogErrNoRows bool

func init() {
	dat.Dialect = postgres.New()
	logger = zap.L().Named("dat:sqlx")
}

// Cache caches query results.
var Cache kvs.KeyValueStore

// SetCache sets this runner's cache. The default cache is in-memory
// based. See cache.MemoryKeyValueStore.
func SetCache(store kvs.KeyValueStore) {
	Cache = store
}

// MustPing pings a database with an exponential backoff. The
// function panics if the database cannot be pinged after 15 minutes
func MustPing(db *sql.DB) {
	var err error
	b := backoff.NewExponentialBackOff()
	ticker := backoff.NewTicker(b)

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for range ticker.C {
		if err = db.Ping(); err != nil {
			logger.Info("pinging database...", zap.Error(err))
			continue
		}

		ticker.Stop()
		return
	}

	panic("Could not ping database!")
}
