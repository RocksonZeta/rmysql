package rmysql

type TimeHash struct {
	Key                 string
	Ttl                 int //seconds
	HeartBeat           int //seconds
	Redis               *RedisService
	MaxNoHeartBeatCount int
}

func NewTimeHash(redis *RedisService, key string, ttl, heartBeat, maxNoHeartBeatCount int) TimeHash {
	return TimeHash{
		Key:                 key,
		Ttl:                 ttl,
		HeartBeat:           heartBeat,
		Redis:               redis,
		MaxNoHeartBeatCount: maxNoHeartBeatCount,
	}
}

//Get 返回key所对应的时间戳，无效<=0。如果存在则，刷新key的时间戳
func (t TimeHash) Get(key string) int {
	var ts int
	t.Redis.HGet(t.Key, key, &ts)
	if ts <= 0 {
		return 0
	}
	t.Redis.HSet(t.Key, key, t.Redis.Now(), int64(t.Ttl))
	return ts
}
func (t TimeHash) Set(key string) {
	t.Redis.HSet(t.Key, key, t.Redis.Now(), int64(t.Ttl))
}
func (t TimeHash) Del(key string) error {
	return t.Redis.HDel(t.Key, key)
}

func (t TimeHash) ListAll() []string {
	m := t.GetAll()
	r := make([]string, len(m))
	i := 0
	for k := range m {
		r[i] = k
		i++
	}
	return r
}
func (t TimeHash) GetAll() map[string]int {
	var r map[string]int
	t.Redis.HMGetAll(t.Key, &r)
	now := int(t.Redis.Now())
	var expiredKeys []string
	for k, v := range r {
		if now-v > t.MaxNoHeartBeatCount*t.HeartBeat {
			expiredKeys = append(expiredKeys, k)
		}
	}
	if len(expiredKeys) > 0 {
		t.Redis.HDel(t.Key, expiredKeys...)
		for _, k := range expiredKeys {
			delete(r, k)
		}
		return r
	}
	return r
}
