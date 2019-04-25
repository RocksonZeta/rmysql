package rmysql_test

import (
	"fmt"
	"rmysql"
	"testing"
	"time"

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
func TestTransaction(t *testing.T) {
	red := rmysql.NewRedisService(conf.Redis)
	defer red.Close()
	red2 := rmysql.NewRedisService(conf.Redis)
	defer red2.Close()
	go func() {
		time.Sleep(time.Second)
		fmt.Println("change k2")
		red2.Set("k2", 2, 10)
		red2.Set("k2", 3, 10)
	}()
	red.Set("k2", 1, 10)
	result := red.WithTransaction(func() {
		red.Set("k1", 1, 10)
		red.Get("k1")
		fmt.Println("set k2")
		red.Set("k2", 1, 10)
		time.Sleep(2 * time.Second)
		red.Get("k2")
	}, "k2")
	fmt.Println(result)
}
