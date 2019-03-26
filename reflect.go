package stcopy

import (
	"errors"
	"reflect"
	"sort"
)

var (
	bytesTyp  = reflect.TypeOf([]byte{})
	stringTyp = reflect.TypeOf("")
)

func NewTypeMap(types []reflect.Type) (r TypeMap) {
	r = TypeMap{}
	for _, val := range types {
		r[val.Name()] = val
	}
	return
}

type TypeMap map[string]reflect.Type

func (m TypeMap) GetKeys() (keys []string) {
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return
}

func (m TypeMap) AddList(list []reflect.Type) {
	for _, val := range list {
		m.Add(val)
	}
}

func (m TypeMap) Add(val reflect.Type) {
	m[val.Name()] = val
}

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

// 如果val是map类型, 检查键值为_type和_ptr的值, 获取反射类型
func (val Value) parseMapType(ctx *Context) (x reflect.Type, err error) {
	srcunfold := val.unfoldInterface()
	if srcunfold.Upper().Kind() != reflect.Map {
		err = errors.New("not map")
		return
	}
	src := val.Upper().Interface().(map[string]interface{})
	// 处理类型
	sttype := func() (y reflect.Type) {
		istr, srcok := src["_type"]
		if srcok == false {
			return
		}
		delete(src, "_type")

		str := istr.(string)
		t, typok := ctx.typeMap[str]
		if typok == false {
			return
		}

		y = t
		return
	}()
	// 处理指针
	isPtr := func() (x bool) {
		istr, ok := src["_ptr"]
		if ok == false {
			return false
		}
		delete(src, "_ptr")
		x = istr.(bool)
		return
	}()

	//
	if sttype == nil {
		x = srcunfold.Upper().Type()
	} else {
		x = sttype
	}
	if isPtr {
		x = reflect.PtrTo(x)
	}
	return
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

// 为map对象添加结构类型 {"_ptr":boolean}
func (val Value) updateMapStructPtrBy(source Value) (err error) {
	indirect := source.Indirect()
	if indirect.Upper().Kind() != reflect.Struct {
		return
	}
	ref := val.Indirect()
	if ref.Upper().Kind() != reflect.Map {
		err = errors.New("not map")
		return
	}

	if source.Upper().Kind() == reflect.Ptr {
		ref.Upper().SetMapIndex(reflect.ValueOf("_ptr"), reflect.ValueOf(true))
	}
	return
}

func (val Value) updateMapPtrBy(source Value) (err error) {
	ref := val.Indirect()
	if ref.Upper().Kind() != reflect.Map {
		err = errors.New("not map")
		return
	}

	if source.Upper().Kind() == reflect.Ptr {
		ref.Upper().SetMapIndex(reflect.ValueOf("_ptr"), reflect.ValueOf(true))
	}
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
		a = reflect.ValueOf(float64(valref.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		a = reflect.ValueOf(float64(valref.Uint()))
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
	TypeUtiler = typeUtil(0)
)

type typeUtil int

// 获取正确的反射对象，如果nil，创建新的
func (*typeUtil) UnfoldType(typ reflect.Type) reflect.Type {
	switch typ.Kind() {
	case reflect.Struct:
	case reflect.Ptr:
		typ = typ.Elem()
		return typ
	}

	return typ
}

// 获取
func (sv *typeUtil) GetFieldRecursion(typ reflect.Type) (r []*reflect.StructField) {
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
