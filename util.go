//  Copyright (c) 2015 Couchbase, Inc.

package gofast

import "reflect"
import "unsafe"
import "strings"
import "bytes"
import "fmt"

func bytes2str(bytes []byte) string {
	if bytes == nil {
		return ""
	}
	sl := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	st := &reflect.StringHeader{Data: sl.Data, Len: sl.Len}
	return *(*string)(unsafe.Pointer(st))
}

func str2bytes(str string) []byte {
	if str == "" {
		return nil
	}
	st := (*reflect.StringHeader)(unsafe.Pointer(&str))
	sl := &reflect.SliceHeader{Data: st.Data, Len: st.Len, Cap: st.Len}
	return *(*[]byte)(unsafe.Pointer(sl))
}

func hasString(str string, strs []string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}

func csv2strings(line string, out []string) []string {
	for _, str := range strings.Split(line, ",") {
		if s := strings.Trim(str, " \n\t\r"); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func msgfactory(msg Message) func() interface{} {
	return func() interface{} {
		typeOfMsg := reflect.ValueOf(msg).Elem().Type()
		return reflect.New(typeOfMsg).Interface()
	}
}

func getStackTrace(skip int, stack []byte) string {
	var buf bytes.Buffer
	lines := strings.Split(string(stack), "\n")
	for _, call := range lines[skip*2:] {
		buf.WriteString(fmt.Sprintf("%s\n", call))
	}
	return buf.String()
}

// CborMap2golangMap transforms [][2]interface{} to map[string]interface{}
// that is required for converting golang to cbor and vice-versa.
//func CborMap2golangMap(value interface{}) interface{} {
//	switch items := value.(type) {
//	case []interface{}:
//		for i, item := range items {
//			items[i] = CborMap2golangMap(item)
//		}
//		return items
//	case [][2]interface{}:
//		m := make(map[string]interface{})
//		for _, item := range items {
//			m[item[0].(string)] = CborMap2golangMap(item[1])
//		}
//		return m
//	}
//	return value
//}

// GolangMap2cborMap transforms map[string]interface{} to [][2]interface{}
// that is required for converting golang to cbor and vice-versa.
//func GolangMap2cborMap(value interface{}) interface{} {
//	switch items := value.(type) {
//	case []interface{}:
//		for i, item := range items {
//			items[i] = GolangMap2cborMap(item)
//		}
//		return items
//	case map[string]interface{}:
//		sl := make([][2]interface{}, 0, len(items))
//		for k, v := range items {
//			sl = append(sl, [2]interface{}{k, GolangMap2cborMap(v)})
//		}
//		return sl
//	}
//	return value
//}
