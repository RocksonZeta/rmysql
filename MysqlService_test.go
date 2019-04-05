package rmysql_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"rmysql"
	"testing"
)

type Conf struct {
	Mysql, Redis string
}

var conf Conf

func init() {
	bs, err := ioutil.ReadFile("test.conf")
	if err != nil {
		fmt.Println("no test.conf in cwd")
		return
	}
	err = json.Unmarshal(bs, &conf)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(conf)
}

// var conStr = "admin:sZlryBOgLxAuAtx9@tcp(test.iqidao.com:43120)/iqidao2?charset=utf8mb4"

func TestSelect(t *testing.T) {
	my := rmysql.NewMysqlService(conf.Mysql, "test", true)
	defer my.Close()
	r := my.SelectInt("select count(*) from User where id>?", 100)
	fmt.Println(r)
}
