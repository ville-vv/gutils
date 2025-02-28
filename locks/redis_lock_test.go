package locks

import (
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedisLock_Lock(t *testing.T) {
	rds := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	lc := NewRedisLock(rds)
	key := "Order0001"
	sum := 0
	sumCh := 0
	for i := 0; i < 100; i++ {
		sum += i
	}
	for i := 0; i < 100; i++ {
		go func(a int) {
			unlockFlow, err := lc.Lock(key, time.Second*1)
			if err != nil {
				assert.NoError(t, err)
				return
			}
			sumCh += a
			_ = lc.UnLock(key, unlockFlow)
		}(i)
	}
	time.Sleep(3 * time.Second)
	assert.Equal(t, sum, sumCh)
}

func TestRedisLock_Lock02(t *testing.T) {
	rds := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	lc := NewRedisLock(rds)
	key := "Order0001"
	sum := 0
	sumCh := 0
	for i := 0; i < 100; i++ {
		sum += i
	}
	for i := 0; i < 100; i++ {
		go func(a int) {
			_, err := lc.Lock(key, time.Second*1)
			//fmt.Println("获得锁", a)
			if err != nil {
				assert.NoError(t, err)
				return
			}
			sumCh += a
			//_ = lc.UnLock(lockKey, unlockFlow)
		}(i)
	}
	time.Sleep(6 * time.Second)
	assert.NotEqual(t, sum, sumCh)
}

func BenchmarkRedisLock_Lock(b *testing.B) {
	b.Skip()
	b.StopTimer()
	rds := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	lc := NewRedisLock(rds)
	key := "Order0001"
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		unlockFlow, _ := lc.Lock(key, time.Second*1)
		//fmt.Println(i, unlockFlow)
		_ = lc.UnLock(key, unlockFlow)
	}
}

// BenchmarkRedisLock_Lock-4   	   23115	     52983 ns/op
// BenchmarkRedisLock_Lock-4   	   22861	     54408 ns/op
// BenchmarkRedisLock_Lock-4   	   23173	     52743 ns/op
