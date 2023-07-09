package util

import (
	"reflect"
	"unsafe"
)

func MergeParam(origin *map[string]string, delta map[string]string) {
	if len(delta) == 0 {
		return
	}

	if *origin == nil {
		*origin = delta
		return
	}

	for key, value := range delta {
		(*origin)[key] = value
	}
}

func Byte2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func String2Byte(str string) []byte {
	strHeader := *(*reflect.StringHeader)(unsafe.Pointer(&str))
	sliceHeader := reflect.SliceHeader{
		Data: strHeader.Data,
		Len:  strHeader.Len,
		Cap:  strHeader.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&sliceHeader))
}
