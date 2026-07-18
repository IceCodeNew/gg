package common

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strings"
)

func Base64StdDecode(s string) (string, error) {
	s = strings.TrimSpace(s)
	saver := s
	if len(s)%4 > 0 {
		s += strings.Repeat("=", 4-len(s)%4)
	}
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return saver, err
	}
	return string(raw), nil
}

func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func StringToBool(s string) bool {
	switch strings.ToLower(s) {
	case "true", "yes", "1", "y":
		return true
	default:
		return false
	}
}

func Base64URLDecode(s string) (string, error) {
	s = strings.TrimSpace(s)
	saver := s
	if len(s)%4 > 0 {
		s += strings.Repeat("=", 4-len(s)%4)
	}
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return saver, err
	}
	return string(raw), nil
}

func ObjectToKV(v any, tagName string) (kv []string) {
	value := reflect.ValueOf(v)
	for value.IsValid() && value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return nil
	}
	return appendStructKV(nil, value, tagName, "")
}

func appendStructKV(kv []string, value reflect.Value, tagName, prefix string) []string {
	typeOfValue := value.Type()
	for i := 0; i < value.NumField(); i++ {
		fieldType := typeOfValue.Field(i)
		if !fieldType.IsExported() {
			continue
		}

		name := fieldType.Name
		if tagName != "" {
			tag := strings.Split(fieldType.Tag.Get(tagName), ",")
			if tag[0] == "-" {
				continue
			}
			if tag[0] != "" {
				name = tag[0]
			}
			if slices.Contains(tag[1:], "omitempty") && value.Field(i).IsZero() {
				continue
			}
		}

		key := name
		if prefix != "" {
			key = prefix + "." + name
		}
		fieldValue := value.Field(i)
		if fieldValue.Kind() == reflect.Struct {
			kv = appendStructKV(kv, fieldValue, tagName, key)
			continue
		}
		kv = append(kv, fmt.Sprintf("%s=%v", key, fieldValue.Interface()))
	}
	return kv
}

func MapToKV(m map[string]any) (kv []string) {
	val := reflect.ValueOf(m)
	keys := val.MapKeys()
	for _, k := range keys {
		v := val.MapIndex(k)
		switch v := v.Interface().(type) {
		case map[string]any:
			subKV := MapToKV(v)
			for _, s := range subKV {
				kv = append(kv, fmt.Sprintf("%v.%v", k.String(), s))
			}
		default:
			kv = append(kv, fmt.Sprintf("%v=%v", k.String(), v))
		}
	}
	return kv
}

func StringsToSet(s []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}

func StringsMapToSet(s []string, mapper func(s string) string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, v := range s {
		m[mapper(v)] = struct{}{}
	}
	return m
}

func SliceUint64toUint32(from []uint64) (to []uint32) {
	to = make([]uint32, len(from)*2)
	for i := range from {
		to[i*2+1] = uint32(from[i] & 0xffffffff)
		to[i*2] = uint32((from[i] & 0xffffffff00000000) >> 32)
	}
	return to
}

func SetValue(values *url.Values, key string, value string) {
	if value == "" {
		return
	}
	values.Set(key, value)
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MustGetMapKeys(m any) (keys []string) {
	v := reflect.ValueOf(m)
	vKeys := v.MapKeys()
	for _, k := range vKeys {
		keys = append(keys, k.String())
	}
	return keys
}
