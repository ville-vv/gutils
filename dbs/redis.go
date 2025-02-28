package dbs

import (
	"context"
	"crypto/tls"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	Tls      bool   `json:"tls"`
}

type IRedisDB interface {
	redis.Cmdable
	redis.BitMapCmdable
}

type RedisDB struct {
	rds IRedisDB
}

func InitRedisDB(rds IRedisDB) *RedisDB {
	return &RedisDB{rds: rds}
}

func MustRedisClient(cfg RedisConfig) *redis.Client {
	var tlsConfig *tls.Config
	if cfg.Tls {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	rds := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   5,
		MinIdleConns: 5,
		TLSConfig:    tlsConfig,
	})
	_, err := rds.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	return rds
}

func (sel *RedisDB) ZAddWithMaxSizeEx(key string, score int64, val string, maxSize int, exp int) error {
	luaScript := `
redis.call("ZADD", KEYS[1], ARGV[1], ARGV[2])
local size = redis.call("ZCARD", KEYS[1])
if size > tonumber(ARGV[3]) then
	redis.call("ZREMRANGEBYRANK", KEYS[1], 0, size-tonumber(ARGV[3])-1)
end
return redis.call("EXPIRE", KEYS[1], ARGV[4])
`
	_, err := sel.Eval(luaScript, []string{key}, score, val, maxSize, exp)
	return err
}

func (sel *RedisDB) LPushWithMaxSizeEx(key string, val string, maxSize int, exp int) error {
	luaScript := `
redis.call("LPUSH", KEYS[1], ARGV[1])
redis.call("LTRIM", KEYS[1], 0, tonumber(ARGV[2])-1)
return redis.call("EXPIRE", KEYS[1], ARGV[3])
`
	_, err := sel.Eval(luaScript, []string{key}, val, maxSize, exp)
	return err
}

// HSetEx 设置值并制定过期时间
func (sel *RedisDB) HSetEx(key, field, value string, exp int) error {
	luaScript := `
redis.call("HSET", KEYS[1], ARGV[1], ARGV[2])
return redis.call("EXPIRE", KEYS[1], ARGV[3])
`
	_, err := sel.Eval(luaScript, []string{key}, field, value, exp)
	return err
}

// HMSetEx 设置多个值并指定过期时间
func (sel *RedisDB) HMSetEx(key string, fieldValues map[string]string, exp int) error {
	// Lua script to set multiple field-value pairs and set the expiration time
	luaScript := `
for i=1, #ARGV-1, 2 do
   redis.call("HSET", KEYS[1], ARGV[i], ARGV[i+1])
end
return redis.call("EXPIRE", KEYS[1], ARGV[#ARGV])
`
	args := make([]interface{}, 0, 2*len(fieldValues)+1)
	for field, value := range fieldValues {
		args = append(args, field, value)
	}
	args = append(args, exp)
	_, err := sel.Eval(luaScript, []string{key}, args...)
	return err
}

func (sel *RedisDB) Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	return sel.rds.Eval(context.Background(), script, keys, args...).Result()
}

func (sel *RedisDB) HSet(key, field, value string) (int64, error) {
	return sel.rds.HSet(context.Background(), key, field, value).Result()
}

// HDel 删除哈希中的指定字段
func (sel *RedisDB) HDel(key string, fields ...string) (int64, error) {
	return sel.rds.HDel(context.Background(), key, fields...).Result()
}

// HExists 检查哈希中的字段是否存在
func (sel *RedisDB) HExists(key, field string) (bool, error) {
	return sel.rds.HExists(context.Background(), key, field).Result()
}

// HGet 获取哈希中指定字段的值
func (sel *RedisDB) HGet(key, field string) (string, error) {
	return sel.rds.HGet(context.Background(), key, field).Result()
}

// HGetAll 获取哈希中的所有键值对
func (sel *RedisDB) HGetAll(key string) (map[string]string, error) {
	return sel.rds.HGetAll(context.Background(), key).Result()
}

// HIncrBy 将哈希中的字段值增加指定整数
func (sel *RedisDB) HIncrBy(key, field string, incr int64) (int64, error) {
	return sel.rds.HIncrBy(context.Background(), key, field, incr).Result()
}

// HIncrByFloat 将哈希中的字段值增加指定浮点数
func (sel *RedisDB) HIncrByFloat(key, field string, incr float64) (float64, error) {
	return sel.rds.HIncrByFloat(context.Background(), key, field, incr).Result()
}

// HKeys 获取哈希中的所有字段名
func (sel *RedisDB) HKeys(key string) ([]string, error) {
	return sel.rds.HKeys(context.Background(), key).Result()
}

// HLen 获取哈希中的字段数
func (sel *RedisDB) HLen(key string) (int64, error) {
	return sel.rds.HLen(context.Background(), key).Result()
}

// HMGet 获取哈希中的多个字段值
func (sel *RedisDB) HMGet(key string, fields ...string) ([]interface{}, error) {
	return sel.rds.HMGet(context.Background(), key, fields...).Result()
}

// HMSet 批量设置哈希中的字段值（已废弃，可以直接使用 HSet）
func (sel *RedisDB) HMSet(key string, values ...interface{}) (int64, error) {
	return sel.rds.HSet(context.Background(), key, values...).Result()
}

// HSetNX 如果字段不存在则设置字段的值
func (sel *RedisDB) HSetNX(key, field string, value interface{}) (bool, error) {
	return sel.rds.HSetNX(context.Background(), key, field, value).Result()
}

// HScan 通过游标扫描哈希中的字段及值
func (sel *RedisDB) HScan(key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return sel.rds.HScan(context.Background(), key, cursor, match, count).Result()
}

// HScanNoValues 通过游标扫描哈希中的字段，不返回值
func (sel *RedisDB) HScanNoValues(key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return sel.rds.HScanNoValues(context.Background(), key, cursor, match, count).Result()
}

// Append 向字符串值追加内容
func (sel *RedisDB) Append(key, value string) (int64, error) {
	return sel.rds.Append(context.Background(), key, value).Result()
}

// Decr 将字符串的值减少 1
func (sel *RedisDB) Decr(key string) (int64, error) {
	return sel.rds.Decr(context.Background(), key).Result()
}

// DecrBy 将字符串的值减少指定的整数
func (sel *RedisDB) DecrBy(key string, decrement int64) (int64, error) {
	return sel.rds.DecrBy(context.Background(), key, decrement).Result()
}

// Get 获取字符串值
func (sel *RedisDB) Get(key string) (string, error) {
	return sel.rds.Get(context.Background(), key).Result()
}

// GetRange 获取字符串的指定范围
func (sel *RedisDB) GetRange(key string, start, end int64) (string, error) {
	return sel.rds.GetRange(context.Background(), key, start, end).Result()
}

// GetSet 设置字符串值并返回旧值
func (sel *RedisDB) GetSet(key string, value interface{}) (string, error) {
	return sel.rds.GetSet(context.Background(), key, value).Result()
}

// GetEx 获取字符串值并设置过期时间
func (sel *RedisDB) GetEx(key string, expiration time.Duration) (string, error) {
	return sel.rds.GetEx(context.Background(), key, expiration).Result()
}

// GetDel 获取并删除字符串值
func (sel *RedisDB) GetDel(key string) (string, error) {
	return sel.rds.GetDel(context.Background(), key).Result()
}

// Incr 将字符串的值增加 1
func (sel *RedisDB) Incr(key string) (int64, error) {
	return sel.rds.Incr(context.Background(), key).Result()
}

// IncrBy 将字符串的值增加指定的整数
func (sel *RedisDB) IncrBy(key string, value int64) (int64, error) {
	return sel.rds.IncrBy(context.Background(), key, value).Result()
}

// IncrByFloat 将字符串的值增加指定的浮点数
func (sel *RedisDB) IncrByFloat(key string, value float64) (float64, error) {
	return sel.rds.IncrByFloat(context.Background(), key, value).Result()
}

// MGet 获取多个字符串值
func (sel *RedisDB) MGet(keys ...string) ([]interface{}, error) {
	return sel.rds.MGet(context.Background(), keys...).Result()
}

// MSet 设置多个键值对
func (sel *RedisDB) MSet(values ...interface{}) (string, error) {
	return sel.rds.MSet(context.Background(), values...).Result()
}

// MSetNX 设置多个键值对，只有在所有键都不存在时成功
func (sel *RedisDB) MSetNX(values ...interface{}) (bool, error) {
	return sel.rds.MSetNX(context.Background(), values...).Result()
}

// SetEx 设置字符串值并指定过期时间
func (sel *RedisDB) SetEx(key string, value interface{}, exp int) (string, error) {
	return sel.rds.SetEx(context.Background(), key, value, time.Duration(exp)*time.Second).Result()
}

// SetNX 如果键不存在则设置字符串值
func (sel *RedisDB) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return sel.rds.SetNX(context.Background(), key, value, expiration).Result()
}

// SetXX 如果键存在则设置字符串值
func (sel *RedisDB) SetXX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return sel.rds.SetXX(context.Background(), key, value, expiration).Result()
}

// SetRange 设置字符串的某个偏移位置的值
func (sel *RedisDB) SetRange(key string, offset int64, value string) (int64, error) {
	return sel.rds.SetRange(context.Background(), key, offset, value).Result()
}

// StrLen 获取字符串值的长度
func (sel *RedisDB) StrLen(key string) (int64, error) {
	return sel.rds.StrLen(context.Background(), key).Result()
}
