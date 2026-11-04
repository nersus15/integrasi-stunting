package satusehat

import (
	postgres "github.com/webcore-go/lib-postgres"
	redis "github.com/webcore-go/lib-redis"
	"github.com/webcore-go/webcore/app/core"
)

// APP_LIBRARIES khusus service "satusehat" — tidak dibagi ke service lain.
var APP_LIBRARIES = map[string]core.LibraryLoader{
	"database:postgres": &postgres.PostgresLoader{},
	"cache:redis":       &redis.RedisLoader{},
}
