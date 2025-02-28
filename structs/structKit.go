package structs

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"reflect"
)

// isEmptyValue checks if a given reflect.Value is the zero value for its type.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice, reflect.Chan:
		return v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	default:
		return false
	}
}

// UpdateStruct updates the fields of dst with non-empty values from src.
func UpdateStruct(dst interface{}, src interface{}) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return fmt.Errorf("destination must be a non-nil pointer")
	}
	return updateStruct(dstVal.Elem(), reflect.ValueOf(src))
}

// updateStruct is the recursive function that performs the actual updating.
func updateStruct(dst, src reflect.Value) error {
	if src.Kind() == reflect.Ptr {
		if src.IsNil() {
			return nil
		}
		src = src.Elem()
	}

	if dst.Kind() != reflect.Struct || src.Kind() != reflect.Struct {
		return fmt.Errorf("both destination and source must be structs or pointers to structs")
	}

	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		dstField := dst.FieldByName(src.Type().Field(i).Name)
		if !dstField.IsValid() || !dstField.CanSet() {
			continue
		}

		if !isEmptyValue(srcField) && isEmptyValue(dstField) {
			dstField.Set(srcField)
		} else if isEmptyValue(srcField) && !isEmptyValue(dstField) {
			continue
		} else if !isEmptyValue(srcField) && !isEmptyValue(dstField) {
			switch srcField.Kind() {
			case reflect.Ptr, reflect.Struct:
				if srcField.Kind() == reflect.Ptr && dstField.Kind() == reflect.Ptr {
					if dstField.IsNil() {
						dstField.Set(reflect.New(srcField.Type().Elem()))
					}
					if err := updateStruct(dstField.Elem(), srcField.Elem()); err != nil {
						return err
					}
				} else if srcField.Kind() == reflect.Struct && dstField.Kind() == reflect.Struct {
					if err := updateStruct(dstField, srcField); err != nil {
						return err
					}
				} else if srcField.Kind() == reflect.Ptr && dstField.Kind() == reflect.Struct {
					if err := updateStruct(dstField, srcField.Elem()); err != nil {
						return err
					}
				} else if srcField.Kind() == reflect.Struct && dstField.Kind() == reflect.Ptr {
					if dstField.IsNil() {
						dstField.Set(reflect.New(srcField.Type()))
					}
					if err := updateStruct(dstField.Elem(), srcField); err != nil {
						return err
					}
				}
			case reflect.Slice:
				dstField.Set(reflect.MakeSlice(srcField.Type(), srcField.Len(), srcField.Cap()))
				reflect.Copy(dstField, srcField)
			case reflect.Map:
				if dstField.IsNil() {
					dstField.Set(reflect.MakeMap(srcField.Type()))
				}
				for _, key := range srcField.MapKeys() {
					dstField.SetMapIndex(key, srcField.MapIndex(key))
				}
			default:
				dstField.Set(srcField)
			}
		}
	}
	return nil
}

// SetValueWithTag 给结构体带有指定tagName字段设置tagName标签后面的值
func SetValueWithTag(to reflect.Value, tagName string) {
	switch to.Kind() {
	case reflect.Struct:
		setFieldValueWithTag(to, tagName)
	default:
	}
}

func setFieldValueWithTag(v reflect.Value, tagName string) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		valueField := v.Field(i)
		// 如果字段是指针且为 nil，则初始化它
		if valueField.Kind() == reflect.Ptr && valueField.IsNil() {
			// 使用 reflect.New 创建一个新的实例
			valueField.Set(reflect.New(valueField.Type().Elem()))
		}
		// 检查是否有默认值标签
		if defaultValue, ok := field.Tag.Lookup(tagName); ok {
			// 如果字段是零值，设置默认值
			if isZero(valueField) {
				// 这里需要处理指针类型的字段
				if valueField.Kind() == reflect.Ptr {
					setFieldValue(valueField.Elem(), defaultValue)
				} else {
					setFieldValue(valueField, defaultValue)
				}
			}
		}
		// 递归处理嵌套结构体
		if valueField.Kind() == reflect.Struct {
			setFieldValueWithTag(valueField, tagName) // 直接递归处理结构体
		} else if valueField.Kind() == reflect.Ptr && valueField.Elem().Kind() == reflect.Struct {
			// 如果是指针类型的结构体，递归处理其内部结构
			setFieldValueWithTag(valueField.Elem(), tagName)
		}
	}
}

// setFieldValue 将默认值设置到字段上
func setFieldValue(field reflect.Value, defaultValue string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(defaultValue)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		field.SetInt(reflect.ValueOf(defaultValue).Int())
	case reflect.Uint, reflect.Uint64:
		field.SetUint(reflect.ValueOf(defaultValue).Uint())
	case reflect.Float32, reflect.Float64:
		field.SetFloat(reflect.ValueOf(defaultValue).Float())
	case reflect.Bool:
		field.SetBool(defaultValue == "true")
	case reflect.Map:
		// 如果字段是 map 类型，则使用 JSON 初始化
		if field.IsZero() {
			// 创建一个空的 map
			field.Set(reflect.MakeMap(field.Type()))
		}
		// 假设 defaultValue 是一个 JSON 字符串
		var mapValue map[string]interface{}
		err := jsoniter.UnmarshalFromString(defaultValue, &mapValue)
		if err == nil {
			for key, val := range mapValue {
				if isZero(field.MapIndex(reflect.ValueOf(key))) {
					field.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
				}
			}
		}
	case reflect.Slice:
		// 如果字段是 slice 类型，则使用 JSON 初始化
		if field.IsZero() {
			// 创建一个空的 slice
			field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		}
		// 假设 defaultValue 是一个 JSON 字符串，表示要初始化的切片
		var sliceValue []interface{}
		err := jsoniter.UnmarshalFromString(defaultValue, &sliceValue)
		if err == nil {
			for _, item := range sliceValue {
				field.Set(reflect.Append(field, reflect.ValueOf(item)))
			}
		}
	default:
	}
}

// isZero 检查字段是否为零值
func isZero(v reflect.Value) bool {
	return v.IsZero()
}
