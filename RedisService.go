package rmysql

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

// Service the Redis service, contains the config and the redis Pool
type RedisService struct {
	// // Connected is true when the Service has already connected
	// Connected bool
	// // Config the redis config for this redis
	Redis redis.Conn
}

func NewService(cli redis.Conn) *RedisService {
	return &RedisService{Redis: cli}
}

func (r *RedisService) Close() {
	err := r.Redis.Close()
	CheckError(err)
}

// PingPong sends a ping and receives a pong, if no pong received then returns false and filled error
func (r *RedisService) PingPong() bool {
	msg, err := r.Redis.Do("PING")
	CheckError(err)
	if err != nil || msg == nil {
		return false
	}
	return (msg == "PONG")
}

func (r *RedisService) AddJsonsId(objs interface{}, keyFn func(id int) string, ttl int64) {
	r.AddJsonsInt(objs, "Id", keyFn, ttl)
}
func (r *RedisService) AddJsonsInt(objs interface{}, field string, keyFn func(id int) string, ttl int64) {
	objsV := reflect.ValueOf(objs)
	objsLen := objsV.Len()
	keys := make([]string, objsLen)
	for i := 0; i < objsLen; i++ {
		keys[i] = keyFn(int(objsV.Index(i).FieldByName(field).Int()))
	}
	r.SetJsons(keys, objs, ttl)
}
func (r *RedisService) SetJson(key string, value interface{}, secondsLifetime int64) (err error) {
	bs := r.marshal(value)
	if secondsLifetime > 0 {
		_, err = r.Redis.Do("SETEX", key, secondsLifetime, bs)
	} else {
		_, err = r.Redis.Do("SET", key, bs)
	}

	return
}

//return (exists , error)
func (r *RedisService) GetJson(key string, out interface{}) bool {
	redisVal, err := r.Redis.Do("GET", key)
	CheckError(err)
	if redisVal == nil {
		return false
	}
	r.unmarshal(redisVal, out)
	return true
}

//GetJsons eg.GetJsons({"1","2"} , &[]obj)
func (r *RedisService) GetJsons(keys []string, out interface{}) error {
	vals, err := redis.Strings(r.Redis.Do("MGET", ArgsString(keys)...))
	CheckError(err)
	if vals == nil {
		return nil
	}
	outT := reflect.TypeOf(out).Elem()
	eleT := outT.Elem()
	nonEmptyLen := 0
	for _, v := range vals {
		if v != "" {
			nonEmptyLen++
		}
	}
	rV := reflect.MakeSlice(outT, nonEmptyLen, nonEmptyLen)
	j := 0
	for i := 0; i < len(keys); i++ {
		if vals[i] != "" {
			ele := reflect.New(eleT)
			eleInterface := ele.Interface()
			err = json.Unmarshal([]byte(vals[i]), &eleInterface)
			CheckError(err)
			rV.Index(j).Set(ele.Elem())
			j++
		}
	}
	reflect.ValueOf(out).Elem().Set(rV)
	return err
}
func (r *RedisService) GetJsonMap(keys []string, out interface{}) error {
	vals, err := redis.Strings(r.Redis.Do("MGET", ArgsString(keys)...))
	CheckError(err)
	if vals == nil {
		return nil
	}
	outT := reflect.TypeOf(out).Elem()
	eleT := outT.Elem()
	rV := reflect.MakeMap(outT)
	for i := 0; i < len(keys); i++ {
		if vals[i] != "" {
			ele := reflect.New(eleT)
			eleInterface := ele.Interface()
			err = json.Unmarshal([]byte(vals[i]), &eleInterface)
			CheckError(err)
			rV.SetMapIndex(reflect.ValueOf(keys[i]), ele.Elem())
		}
	}
	reflect.ValueOf(out).Elem().Set(rV)
	return err
}

func (r *RedisService) SetJsons(keys []string, values interface{}, ttl int64) {
	args := make([]interface{}, len(keys)*2)
	valuesV := PtrValue(reflect.ValueOf(values))

	for i := 0; i < len(args); i += 2 {
		args[i] = keys[i/2]
		args[i+1] = r.marshal(valuesV.Index(i / 2).Interface())
	}
	_, err := r.Redis.Do("MSET", args...)
	CheckError(err)
	if ttl > 0 {
		r.Expires(ttl, keys)
	}

}
func (r *RedisService) GetMap(key string) map[string]interface{} {
	redisVal, err := r.Redis.Do("GET", key)
	CheckError(err)
	if redisVal == nil {
		return nil
	}
	out := make(map[string]interface{})
	r.unmarshal(redisVal, &out)
	return out
}

// Get returns value, err by its key
//returns nil and a filled error if something bad happened.
func (r *RedisService) Set(key interface{}, value interface{}, seconds int) {
	if seconds <= 0 {
		_, err := r.Redis.Do("SET", key, value)
		CheckError(err)
	} else {
		_, err := r.Redis.Do("SETEX", key, seconds, value)
		CheckError(err)
	}

}
func (r *RedisService) Get(key interface{}) RedisValue {
	v, err := r.Redis.Do("GET", key)
	CheckError(err)
	return RedisValue{Value: v}
}

func (r *RedisService) HSet(key, field string, value interface{}, secondsLifetime int64) {
	bs := r.marshal(value)
	_, err := r.Redis.Do("HSET", key, field, bs)
	CheckError(err)
	if secondsLifetime > 0 {
		r.Expire(key, secondsLifetime)
	}
}
func (r *RedisService) HGet(key, field string, result interface{}) {
	bs, err := redis.String(r.Redis.Do("HGET", key, field))
	CheckError(err)
	if bs == "" {
		return
	}
	r.unmarshal(bs, result)
}
func (r *RedisService) HDel(key string, field ...string) error {
	if len(field) <= 0 {
		return nil
	}
	_, err := r.Redis.Do("HDEL", ArgsStringAhead(key, field))
	if nil != err {
		return err
	}
	return nil
}
func (r *RedisService) HGetAsMap(key, field string) map[string]interface{} {
	bs, err := r.Redis.Do("HGET", key, field)
	CheckError(err)
	if bs == nil {
		return nil
	}
	out := make(map[string]interface{})
	r.unmarshal(bs, &out)
	return out
}

// result: *map[string]interface{}
func (r *RedisService) HMGetAll(key string, result interface{}) {
	v, err := r.Redis.Do("HGETALL", key)
	CheckError(err)
	if v == nil {
		return
	}
	vs := v.([]interface{})
	length := len(vs)
	mType := reflect.TypeOf(result).Elem()
	vType := mType.Elem()
	m := reflect.MakeMapWithSize(mType, length/2)
	for i := 0; i < len(vs); i += 2 {
		ele := reflect.New(vType)
		if bs, ok := vs[i+1].([]byte); ok {
			r.unmarshal(bs, ele.Interface())
		}
		m.SetMapIndex(reflect.ValueOf(string(vs[i].([]byte))), ele.Elem())
	}
	reflect.ValueOf(result).Elem().Set(m)
}

func (r *RedisService) HMGet(key string, result interface{}, fields ...string) {
	fs := make([]interface{}, len(fields)+1)
	fs[0] = key
	for i := range fields {
		fs[i+1] = fields[i]
	}
	bs, err := r.Redis.Do("HMGET", fs...)
	CheckError(err)
	if bs == nil {
		return
	}
	vs := bs.([]interface{})
	length := len(fields)
	sliceType := reflect.TypeOf(result).Elem()
	eleType := sliceType.Elem()
	slice := reflect.MakeSlice(sliceType, length, length)
	for i := 0; i < len(vs); i++ {
		ele := reflect.New(eleType)
		if nil != vs[i] {
			r.unmarshal(vs[i].([]byte), ele.Interface())
		}
		slice.Index(i).Set(ele.Elem())
	}
	reflect.ValueOf(result).Elem().Set(slice)
}

func (r *RedisService) HMSet(key string, kvs map[string]interface{}, secondsLifetime int64) {
	args := make([]interface{}, len(kvs)*2+1)
	args[0] = key
	i := 1
	for k, v := range kvs {
		args[i] = k
		value := r.marshal(v)
		args[i+1] = value
		i += 2
	}
	_, err := r.Redis.Do("HMSET", args...)
	CheckError(err)
	if secondsLifetime > 0 {
		r.Expire(key, secondsLifetime)
	}
}

// objs : a list , eg. []xxx
func (r *RedisService) LSet(key string, objs interface{}, secondsLifetime int64) {
	arr := reflect.ValueOf(objs)
	length := arr.Len()
	args := make([]interface{}, length+1)
	args[0] = key
	for i := 1; i <= length; i++ {
		args[i] = r.marshal(arr.Index(i - 1).Interface())
	}
	_, err := r.Redis.Do("RPUSH", args...)
	CheckError(err)
	if secondsLifetime > 0 {
		r.Expire(key, secondsLifetime)
	}

}
func (r *RedisService) LGet(key string, index int, value interface{}) {
	bs, err := r.Redis.Do("LINDEX", key, index)
	CheckError(err)
	if bs1, ok := bs.([]byte); ok {
		if len(bs1) <= 0 {
			return
		}
		r.unmarshal(bs1, value)
	}
}

// result : is a list point ,eg result = &a ; a is var a []Obj
func (r *RedisService) LGetAll(key string, result interface{}) {
	v, err := r.Redis.Do("LRANGE", key, 0, -1)
	CheckError(err)
	if v == nil {
		return
	}
	vs := v.([]interface{})
	length := len(vs)
	sliceType := reflect.TypeOf(result).Elem()
	slice := reflect.MakeSlice(sliceType, length, length)
	elemType := sliceType.Elem()
	for i := 0; i < length; i++ {
		elemValue := reflect.New(elemType)
		r.unmarshal(vs[i].([]byte), elemValue.Interface())
		slice.Index(i).Set(elemValue.Elem())
	}
	reflect.ValueOf(result).Elem().Set(slice)
}
func (r *RedisService) LLen(key string) int {
	v, err := redis.Int(r.Redis.Do("LLEN", key))
	CheckError(err)
	return v
}

func (r *RedisService) Publish(channel string, msg interface{}) (int, error) {
	bs := r.marshal(msg)
	return redis.Int(r.Redis.Do("PUBLISH", channel, bs))
}
func (r *RedisService) Subscribe(channel string) (*redis.PubSubConn, error) {
	// defer c.Close()
	psc := redis.PubSubConn{Conn: r.Redis}
	err := psc.Subscribe(channel)
	if err != nil {
		return nil, err
	}
	return &psc, nil
	// for {
	// 	switch v := psc.Receive().(type) {
	// 	case redis.Message:
	// 		fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
	// 	case redis.Subscription:
	// 		fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
	// 	case error:
	// 		return v
	// 	}
	// }

}

func (r *RedisService) TTL(key string) int64 {
	redisVal, err := r.Redis.Do("TTL", key)
	CheckError(err)
	return redisVal.(int64)
}

func (r *RedisService) Expire(key string, newSecondsLifeTime int64) {
	_, err := r.Redis.Do("EXPIRE", key, newSecondsLifeTime)
	CheckError(err)
}
func (r *RedisService) Expires(seconds int64, keys []string) {
	for _, k := range keys {
		r.Expire(k, seconds)
	}
}
func (r *RedisService) Exists(key string) bool {
	b, err := redis.Bool(r.Redis.Do("EXISTS", key))
	CheckError(err)
	return b

}
func (r *RedisService) AllExist(keys ...string) bool {
	length := len(keys)
	kks := make([]interface{}, length)
	for i, k := range keys {
		kks[i] = k
	}

	count, err := redis.Int(r.Redis.Do("EXISTS", kks...))
	CheckError(err)
	return count == length
}

// func (r *RedisService) marshalString(v interface{}) string {
// 	bs := r.marshal(v)
// 	return string(bs)
// }
func (r *RedisService) marshal(v interface{}) string {
	bs, err := json.Marshal(v)
	CheckError(err)
	return string(bs)
}

func (r *RedisService) unmarshal(v interface{}, out interface{}) {
	if bs, ok := v.(string); ok {
		err := json.Unmarshal([]byte(bs), out)
		CheckError(err)
		return
	}
	if bs, ok := v.([]byte); ok {
		err := json.Unmarshal(bs, out)
		CheckError(err)
		return
	}
	Panic(reflect.TypeOf(v).String()+" type cast error:not []byte or string type", nil)
}

func (r *RedisService) GetBytes(key string) []byte {
	redisVal, err := r.Redis.Do("GET", key)
	bs, err := redis.Bytes(redisVal, err)
	CheckError(err)
	return bs
}

// Delete removes redis entry by specific key
func (r *RedisService) Delete(key string) {
	_, err := r.Redis.Do("DEL", key)
	CheckError(err)
}
func (r *RedisService) Select(db int) {
	_, err := r.Redis.Do("SELECT", strconv.Itoa(db))
	CheckError(err)
}

//Now seconds
func (r *RedisService) Now() int64 {
	s, _ := r.Time()
	return s
}
func (r *RedisService) Time() (int64, int64) {
	times, err := redis.Int64s(r.Redis.Do("TIME"))
	CheckError(err)
	return times[0], times[1]
}

//Multi Unwatch Watch ... Exec/Discard
func (r *RedisService) Multi() {
	_, err := r.Redis.Do("MULTI")
	CheckError(err)
}
func (r *RedisService) Exec() (interface{}, error) {
	return r.Redis.Do("EXEC")
}
func (r *RedisService) Discard() {
	_, err := r.Redis.Do("DISCARD")
	CheckError(err)
}
func (r *RedisService) Watch(keys ...string) {
	_, err := r.Redis.Do("WATCH", ArgsString(keys))
	CheckError(err)
}
func (r *RedisService) Unwatch() {
	_, err := r.Redis.Do("UNWATCH")
	CheckError(err)
}

type RedisValue struct {
	Value interface{}
}

func (r RedisValue) Int() int {
	v, _ := redis.Int(r.Value, nil)
	return v
}
func (r RedisValue) Ints() []int {
	v, _ := redis.Ints(r.Value, nil)
	return v
}
func (r RedisValue) String() string {
	v, _ := redis.String(r.Value, nil)
	return v
}
func (r RedisValue) Strings() []string {
	v, _ := redis.Strings(r.Value, nil)
	return v
}
func (r RedisValue) Int64() int64 {
	v, _ := redis.Int64(r.Value, nil)
	return v
}
func (r RedisValue) Int64s() []int64 {
	v, _ := redis.Int64s(r.Value, nil)
	return v
}
func (r RedisValue) Bool() bool {
	v, _ := redis.Bool(r.Value, nil)
	return v
}
func (r RedisValue) Float64() float64 {
	v, _ := redis.Float64(r.Value, nil)
	return v
}
func (r RedisValue) Float64s() []float64 {
	v, _ := redis.Float64s(r.Value, nil)
	return v
}
