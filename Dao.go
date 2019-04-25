package rmysql

import (
	"reflect"
)

type Dao struct {
	Mysql *MysqlService
	Redis *RedisService
	TTL   int64
}

func NewDao(mysql *MysqlService, redis *RedisService, ttl int64) Dao {
	return Dao{Mysql: mysql, Redis: redis, TTL: ttl}
}

func (d Dao) Close() {
	d.Redis.Close()
	d.Mysql.Close()
}
func (d Dao) Get(id int, result interface{}, key string) {
	d.Redis.GetJson(key, result)
	idValue := GetField(result, "Id").(int)
	if idValue <= 0 {
		d.Mysql.Get(result, id)
		d.Redis.SetJson(key, result, d.TTL)
	}
}

//List  result: eg. &[]table.User
func (d Dao) ListBySql(result interface{}, key string, query string, args ...interface{}) {
	exists := d.Redis.GetJson(key, &result)
	if exists {
		return
	}
	d.Mysql.Select(result, query, args...)
	d.Redis.SetJson(key, result, d.TTL)
}
func (d Dao) List(ids []int, result interface{}, redisKey func(int) string) {
	ids = Unique(ids)
	resultV := reflect.ValueOf(result).Elem()
	// resultT := resultV.Type()
	// fmt.Println(reflect.TypeOf(&result))
	d.Redis.GetJsons(MapIds(ids, redisKey), result)
	if resultV.Len() == len(ids) {
		return
	}
	// inRedisIds := structutil.ColInt(result, "Id")
	// leftIds := stringutil.IntsSubstract(ids, inRedisIds)
	// structutil.MapValues(r, &result)
	// left := reflect.MakeSlice(resultT, 0, 0).
	// left := resultType
	resultV.Set(reflect.Zero(reflect.TypeOf(result).Elem()))
	d.Mysql.List(result, ids)
	d.Redis.SetJsons(ColInt2Str(result, "Id", redisKey), result, d.TTL)
	// result = append(result, left...)
	// structutil.Join(result, left, result)
	return
}
