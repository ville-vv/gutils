package uuids

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/oklog/ulid/v2"
	"math/rand"
	"time"
)

func UUID() string {
	return uuid.New().String()
}

func UUIDInMark(mark string) string {
	namespace := uuid.Must(uuid.NewRandom())
	return uuid.NewSHA1(namespace, []byte(mark)).String()
}

func UUIDBase58() string {
	u := uuid.Must(uuid.NewRandom())
	return base58.Encode(u[:])
}

func NanoID() string {
	id, _ := gonanoid.New()
	return id
}

// ULID 说明 UUID/UUID 在许多使用场景中可能不是最优选择，因为：
// - 它不是对 128 位数据最字符高效的编码方式
// - UUID v1/v2 在许多环境中不实用，因为它需要访问唯一且稳定的 MAC 地址
// - UUID v3/v5 需要一个唯一的种子并生成随机分布的 ID，这会导致许多数据结构中的碎片化
// - UUID v4 仅提供随机性，没有其他信息，这也会在许多数据结构中导致碎片化
// 相比之下，ULID 具有以下优势：
// - 与 UUID/GUID 兼容
// - 每毫秒可生成 1.21e+24 个唯一 ULID（准确数字为 1,208,925,819,614,629,174,706,176）
// - 可按字典顺序排序
// - 编码为 26 个字符的字符串，而 UUID 则是 36 个字符
// - 使用 Crockford 的 base32 编码以提高效率和可读性（每字符 5 位）
// - 不区分大小写
// - 无特殊字符（URL 安全）
// - 单调排序（正确检测并处理相同毫秒生成的 ID）
func ULID() string {
	t := time.Now().UTC()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return id.String()
}

// SnowFlakeID 雪花算法
func SnowFlakeID() string {
	node, err := snowflake.NewNode(562)
	if err != nil {
		panic(err)
	}
	id := node.Generate()
	return id.String()
}
