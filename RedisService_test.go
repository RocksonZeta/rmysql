package rmysql_test

import (
	"rmysql"
	"testing"

	"github.com/stretchr/testify/assert"
)

var host = "test.iqidao.com:50002"

func TestPing(t *testing.T) {
	red := rmysql.NewRedisServices(host, "").Get()
	defer red.Close()
	assert.True(t, red.PingPong())
}
func TestGetSet(t *testing.T) {
	red := rmysql.NewRedisServices(host, "").Get()
	defer red.Close()
	red.Set("k1", 1, 10)
	assert.Equal(t, 1, red.Get("k1").Int())
}
