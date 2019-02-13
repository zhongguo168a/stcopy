package stcopy

import (
	"github.com/pkg/errors"
	"reflect"
)

func NewTypeMap(types []reflect.Type) (r TypeMap) {
	r = TypeMap{}
	for _, val := range types {
		r[val.Name()] = val
	}
	return
}

type TypeMap map[string]reflect.Type

type Value reflect.Value

func (val Value) Upper() reflect.Value {
	return reflect.Value(val)
}

func (val Value) Indirect() Value {
	return Value(reflect.Indirect(val.Upper()))
}

func (val Value) unfoldInterface() (r Value) {
	ref := val.Upper()
	if ref.Kind() == reflect.Interface {
		return Value(ref.Elem())
	}
	return val
}

// 为map对象添加结构类型 {"_type":Name}
func (val Value) updateMapStructTypeBy(source Value) (err error) {
	indirect := source.Indirect()
	if indirect.Upper().Kind() != reflect.Struct {
		//
		return
	}

	ref := val.Indirect()
	if ref.Upper().Kind() != reflect.Map {
		err = errors.New("not map")
		return
	}

	ref.Upper().SetMapIndex(reflect.ValueOf("_type"), reflect.ValueOf(indirect.Upper().Type().Name()))
	return
}

func (val Value) GetTypeString() (y string) {
	if val.Upper().IsValid() {
		y = val.Upper().Type().String()
	} else {
		y = "is nil"
	}
	return
}

func (val Value) IsNil() bool {
	if isHard(val.Upper().Kind()) {
		return val.IsNil()
	}
	return false
}

// 转化成map类型的值
func (val Value) convertToMapValue() (r Value) {
	valref := val.Upper()
	var a reflect.Value
	switch valref.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		a = reflect.ValueOf(int(valref.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		a = reflect.ValueOf(valref.Uint())
	case reflect.Float32, reflect.Float64:
		a = reflect.ValueOf(valref.Float())
	case reflect.Bool:
		a = reflect.ValueOf(valref.Bool())
	case reflect.String:
		a = reflect.ValueOf(valref.String())
	default:
		a = valref
	}
	r = Value(a)
	return
}

var (
	TypeUtiler = TypeUtil(0)
)

type TypeUtil int

// 获取正确的反射对象，如果nil，创建新的
func (*TypeUtil) UnfoldType(typ reflect.Type) reflect.Type {
	switch typ.Kind() {
	case reflect.Struct:
	case reflect.Ptr:
		typ = typ.Elem()
		return typ
	}

	return typ
}

// 获取
func (sv *TypeUtil) GetFieldRecursion(typ reflect.Type) (r []*reflect.StructField) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		switch field.Type.Kind() {
		case reflect.Struct:
			if field.Anonymous == true {

				r = append(r, sv.GetFieldRecursion(field.Type)...)
			} else {
				r = append(r, &field)
			}

		default:
			r = append(r, &field)
		}
	}
	return
}
