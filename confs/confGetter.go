package confCenter

import "time"

type ConfGetter interface {
	Get(key string) any
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetInt(key string) int
	GetInt64(key string) int64
	GetIntSlice(key string) []int
	GetString(key string) string
	GetStringMap(key string) map[string]any
	GetStringMapString(key string) map[string]string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	IsSet(key string) bool
	AllSettings() map[string]any
	GetSubConfig(key string) ConfGetter // Sub not support watch
	Unmarshal(rawVal any) error
	UnmarshalKey(key string, rawVal any) error
}

func WrapDefaultValueGetter(cc ConfGetter) ConfDefaultValueGetter {
	return &getWithDefault{ConfGetter: cc}
}

type ConfDefaultValueGetter interface {
	ConfGetter
	GetStringDefault(key string, def string) string
	GetFloat64Default(key string, def float64) float64
	GetIntDefault(key string, def int) int
	GetInt64Default(key string, def int64) int64
}

type getWithDefault struct {
	ConfGetter
}

func (sel *getWithDefault) GetStringDefault(key string, def string) string {
	res := sel.GetString(key)
	if res != "" {
		return res
	}
	return def
}

func (sel *getWithDefault) GetFloat64Default(key string, def float64) float64 {
	res := sel.GetFloat64(key)
	if !(res < 1e-9 && res > 0) {
		return def
	}
	return def
}

func (sel *getWithDefault) GetIntDefault(key string, def int) int {
	res := sel.GetInt(key)
	if res != 0 {
		return res
	}
	return def
}

func (sel *getWithDefault) GetInt64Default(key string, def int64) int64 {
	res := sel.GetInt64(key)
	if res != 0 {
		return res
	}
	return def
}
