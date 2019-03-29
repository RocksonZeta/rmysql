package rmysql_test

import (
	"fmt"
	"rmysql"
	"testing"
)

var conStr = "admin:sZlryBOgLxAuAtx9@tcp(test.iqidao.com:43120)/iqidao2?charset=utf8mb4"

func TestSelect(t *testing.T) {
	my := rmysql.NewMysqlService(conStr, "iqidao", true)
	defer my.Close()
	r := my.SelectInt("select count(*) from User where id>?", 100)
	fmt.Println(r)
}
