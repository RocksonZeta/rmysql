package rmysql_test

import (
	"rmysql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	red := rmysql.NewRedisService(conf.Redis)
	defer red.Close()
	assert.True(t, red.PingPong())
}

func TestGetSet(t *testing.T) {
	red := rmysql.NewRedisService(conf.Redis)
	defer red.Close()
	red.Set("k1", 1, 10)
	assert.Equal(t, 1, red.Get("k1").Int())
}
