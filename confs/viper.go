package confCenter

import (
	"bytes"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/ville-vv/gutils/structs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)

type Viper struct {
	*viper.Viper
}

func getNameAndExt(fileName string) (string, string) {
	ls := strings.Split(fileName, ".")
	if len(ls) < 2 {
		return ls[0], ""
	}
	if len(ls) == 2 {
		return ls[0], ls[1]
	}
	return strings.Join(ls[:len(ls)-1], "."), ls[len(ls)-1]
}

// NewViper
// fileName 要包含文件名和文件名後綴
// fileType 文件类型为空就会去 fileName 的后缀，如果不为空就以fileType为主要
func NewViper(fileName string) (*Viper, error) {
	v := &Viper{Viper: viper.New()}
	if fileName == "" {
		return v, v.ReadConfig(bytes.NewBufferString(""))
	}
	ext := ""
	fileName, ext = getNameAndExt(fileName)
	v.SetConfigName(path.Base(fileName))
	v.SetConfigType(ext)
	v.AddConfigPath("./etc/")
	v.AddConfigPath("/etc/")
	v.AddConfigPath(".")
	if homeDir, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(homeDir))
	}
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}

func NewViperWithWatch(fileName string) (*Viper, error) {
	v, err := NewViper(fileName)
	if err != nil {
		return nil, err
	}
	v.WatchConfig()
	return v, nil
}

func NewViperWatchToObj(fileName string, obj any) (*Viper, error) {
	v, err := NewViper(fileName)
	if err != nil {
		return nil, err
	}
	// 将配置文件映射到结构体
	if err = v.decodeWithDefaults(&obj); err != nil {
		return nil, err
	}
	v.OnConfigChange(func(in fsnotify.Event) {
		//objVal := reflect.ValueOf(obj)
		//newObj := reflect.New(objVal.Type().Elem()).Interface()
		//if err := v.Unmarshal(&newObj); err != nil {
		//	zlog.Errorw("OnConfigChange unmarshal config file error", "errMsg", err.Error())
		//}
		//// 使用反射将 newObj 的值赋值给 obj
		//objVal.Elem().Set(reflect.ValueOf(newObj).Elem())
	})
	v.WatchConfig()
	return v, nil
}

func MustNewViperWatchToObj(fileName string, obj any) *Viper {
	c, err := NewViperWatchToObj(fileName, obj)
	if err != nil {
		panic(err)
	}
	return c
}

func MustNewViperWithWatch(fileName string) *Viper {
	c, err := NewViperWithWatch(fileName)
	if err != nil {
		panic(err)
	}
	return c
}

func (sel *Viper) GetSubConfig(key string) ConfGetter {
	return &Viper{Viper: sel.Viper.Sub(key)}
}

func (sel *Viper) UnmarshalKey(key string, val interface{}) error {
	return sel.Viper.UnmarshalKey(key, val)
}

func (sel *Viper) Unmarshal(val interface{}) error {
	return sel.Viper.Unmarshal(val)
}

// decodeWithDefaults 通过 viper 解析配置并使用 default 标签设置默认值
func (sel *Viper) decodeWithDefaults(config interface{}) error {
	configDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
		Result:           config,
		DecodeHook: func(from, to reflect.Value) (interface{}, error) {
			structs.SetValueWithTag(to, "default")
			return from.Interface(), nil
		},
	})
	if err != nil {
		return err
	}
	return configDecoder.Decode(sel.AllSettings())
}
