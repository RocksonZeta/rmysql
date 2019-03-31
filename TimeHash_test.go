package rmysql_test

import (
	"fmt"
	"rmysql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeHashGet1(t *testing.T) {
	red := rmysql.NewRedisServices(host, "").Get()
	th := rmysql.NewTimeHash(red, "test-th1", 30, 5, 3)
	defer th.Redis.Close()
	th.Set("k1")
	th.Set("k3")
	v2 := th.Get("k1")
	assert.True(t, v2 > 0)
	v3 := th.GetAll()
	assert.True(t, len(v3) > 0)
	fmt.Println(v3)
	fmt.Println(v3)

}
