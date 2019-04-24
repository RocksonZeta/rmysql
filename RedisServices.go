package rmysql

import (
	"github.com/gomodule/redigo/redis"
)

// type RedisConfig struct {
// 	// Network "tcp"
// 	// Network string
// 	// Addr "127.0.0.1:6379"
// 	// Addr     string
// 	Url string
// 	// Password string
// 	Db int
// 	// <0 no limit
// 	MaxIdle int
// 	// < 0 no limit
// 	MaxActive int
// 	// IdleTimeout 1 * time.Minute
// 	IdleTimeout time.Duration
// }

//NewRedisServices redisUrl like redis://user:secret@localhost:6379/0?foo=bar&qux=baz
func NewRedisService(redisUrl string) *RedisService {
	con, err := redis.DialURL(redisUrl)
	if err != nil {
		panic(err)
	}
	return &RedisService{Redis: con}
}
