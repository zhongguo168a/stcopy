package stcopy

import (
	"reflect"
	"strconv"
)

func convert2String(val reflect.Value, typ reflect.Type) (r reflect.Value) {

	switch val.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		r = reflect.ValueOf(strconv.Itoa(int(val.Uint())))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		r = reflect.ValueOf(strconv.Itoa(int(val.Int())))
	case reflect.Float32, reflect.Float64:
		r = reflect.ValueOf(strconv.FormatFloat(val.Float(), 'f', -1, 64))
	default:
		if val.Type().ConvertibleTo(typ) {
			r = val.Convert(typ)
		} else {
			r = reflect.ValueOf("")
		}

	}
	return
}

// 转换成Int型
func convert2Int(val reflect.Value, typ reflect.Type) (r reflect.Value) {
	switch val.Kind() {
	case reflect.String:
		i, err := strconv.Atoi(val.Interface().(string))
		if err != nil {
			r = reflect.ValueOf(0)
		} else {
			r = reflect.ValueOf(i)
		}
	default:
		if val.Type().ConvertibleTo(typ) {
			r = val
		} else {
			r = reflect.ValueOf(0)
		}
	}

	r = r.Convert(typ)
	return
}

func convert2Float(val reflect.Value, typ reflect.Type) (r reflect.Value) {
	switch val.Kind() {
	case reflect.String:
		i, err := strconv.ParseFloat(val.Interface().(string), 64)
		if err != nil {
			r = reflect.ValueOf(0.0)
		} else {
			r = reflect.ValueOf(i)
		}
	default:
		if val.Type().ConvertibleTo(typ) {
			r = val
		} else {
			r = reflect.ValueOf(0.0)
		}
	}

	r = r.Convert(typ)
	return
}

func convert2Bool(val reflect.Value, typ reflect.Type) (r reflect.Value) {
	switch val.Kind() {
	case reflect.String:
		data := val.Interface().(string)
		if data == "true" || data == "1" {
			r = reflect.ValueOf(true)
		} else {
			r = reflect.ValueOf(false)
		}
	default:
		if val.Type().ConvertibleTo(typ) {
			r = val
		} else {
			r = reflect.ValueOf(false)
		}
	}

	r = r.Convert(typ)
	return
}
