package stcopy

import (
	"errors"
	"reflect"
)

type Description struct {
	// 标签
	Tags map[string]string
	// 类型
	Typ reflect.Type
	// 枚举
	Enum *Enum
	// 描述集合
	descs DescriptionMap
}

func (desc *Description) IsEnum() bool {
	switch desc.Typ.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
	case reflect.String:
		return true
	}

	return false
}

func (desc *Description) Contains(tag string) bool {
	if desc.Tags == nil {
		return false
	}
	_, ok := desc.Tags[tag]
	return ok
}

func (desc *Description) GetEnumValue(key string) (val interface{}, err error) {
	val, ok := desc.Enum.Name2Value[key]
	if ok == false {
		err = errors.New("not found")
	}
	return
}

func (desc *Description) GetEnumKeys() (r []reflect.Value) {
	for _, val := range desc.Enum.Names() {
		r = append(r, reflect.ValueOf(val))
	}
	return
}

func (desc *Description) GetParent() (r reflect.Type) {
	if desc.Typ.NumField() == 0 {
		return
	}
	f := desc.Typ.Field(0)
	if f.Anonymous == false {
		return
	}

	r = f.Type
	return
}

func (desc *Description) GetFieldRecursion() (r []*reflect.StructField) {
	typ := desc.Typ
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		switch field.Type.Kind() {
		case reflect.Struct:
			if field.Anonymous == true {
				parent, ok := desc.descs[field.Type.Name()]
				if ok == false {
					println("not description:", field.Type.Name())
					continue
				}
				r = append(r, parent.GetFieldRecursion()...)
			} else {
				r = append(r, &field)
			}

		default:
			r = append(r, &field)
		}
	}
	return
}

func (desc *Description) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"name": desc.Typ.Name(),
		"type": func() (x string) {
			switch desc.Typ.Kind() {
			case reflect.Slice:
				fallthrough
			case reflect.Array:
				x = "array"
			case reflect.Map:
				x = "map"
			case reflect.Struct:
				x = "struct"
			default:
				x = desc.Typ.Kind().String()
			}
			return
		}(),
		"item": func() (x string) {
			switch desc.Typ.Kind() {
			case reflect.Slice:
				fallthrough
			case reflect.Array:
				x = "array"
			case reflect.Map:
				x = "map"
			}
			return
		}(),
		"struct": func() (x string) {
			switch desc.Typ.Kind() {
			case reflect.Slice:
				fallthrough
			case reflect.Array:
				fallthrough
			case reflect.Map:
				itemKind := desc.Typ.Elem().Kind()
				switch itemKind {
				case reflect.Struct:
					x = desc.Typ.Elem().String()
				}
			case reflect.Struct:
				x = desc.Typ.String()
			}
			return
		}(),
		"enum": desc.IsEnum(),
		"custom": func() (x string) {
			return
		}(),
	}
}
