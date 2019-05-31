package stcopy

import (
	"reflect"
	"strconv"
)

func convert2String(val reflect.Value, typ reflect.Type) (r reflect.Value) {
	switch val.Kind() {
	default:
		if val.Type().ConvertibleTo(typ) {
			r = val
		} else {
			r = reflect.ValueOf("")
		}
	}
	r = r.Convert(typ)
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

func convert2Bool(val reflect.Value, typ reflect.Type) (r reflect.Value) {
	switch val.Kind() {
	case reflect.String:
		data := val.Interface().(string)
		if data == "true" || data == "1" {
			val = reflect.ValueOf(true)
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
