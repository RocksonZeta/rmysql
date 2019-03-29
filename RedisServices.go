package rmysql

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

type RedisConfig struct {
	// Network "tcp"
	Network string
	// Addr "127.0.0.1:6379"
	Addr     string
	Password string
	Db       int
	// <0 no limit
	MaxIdle int
	// < 0 no limit
	MaxActive int
	// IdleTimeout 1 * time.Minute
	IdleTimeout time.Duration
}

type RedisServices struct {
	Config RedisConfig
	Pool   *redis.Pool
}

func NewRedisServices(addr, password string) *RedisServices {

	r := &RedisServices{Config: RedisConfig{Addr: addr, Password: password}}
	r.initPool()
	return r
}
func RedisServicesWithConfig(Config RedisConfig) *RedisServices {
	r := &RedisServices{Config: Config}
	r.initPool()
	return r
}

func (s *RedisServices) Get() *RedisService {
	return &RedisService{Redis: s.Pool.Get()}
}
func (s *RedisServices) GetDb(db int) *RedisService {
	r := &RedisService{Redis: s.Pool.Get()}
	r.Select(db)
	return r
}

// Connect connects to the redis, called only once
func (s *RedisServices) initPool() {
	c := s.Config

	if c.IdleTimeout <= 0 {
		c.IdleTimeout = 60 * time.Second
	}

	if c.Network == "" {
		c.Network = "tcp"
	}

	if c.Addr == "" {
		c.Addr = "127.0.0.1:6379"
	}
	if c.MaxIdle == 0 {
		c.MaxIdle = 3
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 1 * time.Minute
	}

	Pool := &redis.Pool{IdleTimeout: c.IdleTimeout, MaxIdle: c.MaxIdle, MaxActive: c.MaxActive}
	Pool.TestOnBorrow = func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	}

	Pool.Dial = func() (redis.Conn, error) {
		return dial(c)
	}
	s.Pool = Pool
}

func dial(config RedisConfig) (redis.Conn, error) {
	var opts []redis.DialOption
	if config.Db != 0 {
		opts = append(opts, redis.DialDatabase(config.Db))
	}
	if config.Password != "" {
		opts = append(opts, redis.DialPassword(config.Password))
	}
	c, err := redis.Dial(config.Network, config.Addr, opts...)
	return c, err
}
