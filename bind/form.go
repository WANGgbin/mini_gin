package bind

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

type formBinder struct {}

func (f *formBinder) Bind(req *http.Request, target interface{}) error {
	err := req.ParseForm()
	if err != nil {
		return err
	}

	return doBind(req.Form, reflect.ValueOf(target))
}

func doBind(input url.Values, target reflect.Value) error {
	elemTyp := target.Type().Elem()
	elemVal := target.Elem()
	switch elemTyp.Kind() {
	case reflect.Struct:
		fmt.Printf("[before] name: %s\n", elemTyp.Name())
		return bindUrlValuesToStruct(input, elemVal)
	case reflect.Ptr:
		if elemVal.IsNil() {
			fmt.Printf("name: %s\n", elemTyp.Elem().Name())
			newVal := reflect.New(elemTyp.Elem())
			elemVal.Set(newVal)
		}
		return doBind(input, elemVal)
	default:
		return ErrUnknownType
	}
}

var (
	ErrUnknownType = errors.New("unknown type")
)

func bindUrlValuesToStruct(input url.Values, targetVal reflect.Value) error {
	targetTyp := targetVal.Type()

	// target 字段类型只能是 bool/int/float/string 即对应的 array/slice 格式
	for idx := 0; idx < targetTyp.NumField(); idx++ {
		fieldTyp := targetTyp.Field(idx)
		fieldVal := targetVal.Field(idx)

		// 未导出字段无法赋值，需要跳过，负责 panic !!!
		if fieldTyp.PkgPath != "" {
			continue
		}

		tag := parseTag(fieldTyp.Tag.Get("form"))
		if  tag.name == ""{
			tag.name = fieldTyp.Name
		}
		vals, exist := input[tag.name]
		if !exist {
			if tag.setDefault {
				vals = []string{tag.defaultVal}
			} else {
				continue
			}
		}

		switch fieldTyp.Type.Kind() {
		case reflect.Array:
			if fieldTyp.Type.Len() != len(vals) {
				return fmt.Errorf("length of input is not equal to length of field %s", fieldTyp.Name)
			}

			for idx := 0; idx < len(vals); idx++ {
				err := bindStringToValue(vals[idx], fieldVal.Index(idx), fieldTyp)
				if err != nil {
					return err
				}
			}
		case reflect.Slice:
			newSlice := reflect.MakeSlice(fieldTyp.Type, len(vals), len(vals))
			for idx := 0; idx < newSlice.Len(); idx++ {
				err := bindStringToValue(vals[idx], newSlice.Index(idx), fieldTyp)
				if err != nil {
					return err
				}
			}
			fieldVal.Set(newSlice)
		default:
			err := bindStringToValue(vals[0], fieldVal, fieldTyp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func bindStringToValue(val string, target reflect.Value, fieldType reflect.StructField) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parse value [%s] to field [%s] error: %v", val, fieldType.Name, err)
		}
	}()
	switch target.Kind() {
	case reflect.String:
		target.SetString(val)
		return nil
	case reflect.Bool:
		return bindBool(target, val)
	case reflect.Int8:
		return bindInt(target, val, 8)
	case reflect.Int16:
		return bindInt(target, val, 16)
	case reflect.Int:
		return bindInt(target, val, 8*int(unsafe.Sizeof(0)))
	case reflect.Int32:
		return bindInt(target, val, 32)
	case reflect.Int64:
		return bindInt(target, val, 64)
	case reflect.Uint8:
		return bindUint(target, val, 8)
	case reflect.Uint16:
		return bindUint(target, val, 16)
	case reflect.Uint:
		return bindUint(target, val, 8*int(unsafe.Sizeof(0)))
	case reflect.Uint32:
		return bindUint(target, val, 32)
	case reflect.Uint64:
		return bindUint(target, val, 64)
	case reflect.Float32:
		return bindFloat(target, val, 32)
	case reflect.Float64:
		return bindFloat(target, val, 64)
	default:
		return ErrUnknownType
	}
}

func bindInt(fieldVal reflect.Value, val string, bitSize int) error {
	num, err := strconv.ParseInt(val, 10, bitSize)
	if err != nil {
		return err
	}
	fieldVal.SetInt(num)
	return nil
}

func bindUint(fieldVal reflect.Value, val string, bitSize int) error {
	num, err := strconv.ParseUint(val, 10, bitSize)
	if err != nil {
		return err
	}
	fieldVal.SetUint(num)
	return nil
}

func bindFloat(fieldVal reflect.Value, val string, bitSize int) error {
	num, err := strconv.ParseFloat(val, bitSize)
	if err != nil {
		return err
	}
	fieldVal.SetFloat(num)
	return nil
}

func bindBool(fieldVal reflect.Value, val string) error {
	num, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	fieldVal.SetBool(num)
	return nil
}

// formTag "name,default=value"
type formTag struct {
	name string

	setDefault bool
	defaultVal string
}

func parseTag(tag string) *formTag {
	segs := strings.Split(tag, ",")
	if len(segs) == 0 {
		return nil
	}
	ret := new(formTag)
	for _, seg := range segs {
		idx := strings.Index(seg, "=")
		if idx == -1 {
			ret.name = seg
		} else {
			key := seg[:idx]
			val := seg[idx+1:]
			if key == "default" {
				ret.setDefault = true
				ret.defaultVal = val
			}
		}
	}
	return ret
}