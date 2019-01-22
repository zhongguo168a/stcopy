package stcopy

import (
	"errors"
	"reflect"
)

type DescriptionMap map[string]*Description

func (m DescriptionMap) AddList(typs []*Description) {
	for _, val := range typs {
		m.Add(val)

	}
}

func (m DescriptionMap) CreateByTyp(typ reflect.Type) (desc *Description) {
	desc = &Description{
		Typ: typ,
	}
	m.Add(desc)
	return
}

func (m DescriptionMap) Add(typ *Description) {
	typ.descs = m
	m[typ.Typ.Name()] = typ
	m[typ.Typ.String()] = typ
}

func (m DescriptionMap) GetBy(key string) (*Description, error) {
	desc, ok := m[key]
	if ok == false {
		return nil, errors.New("not found[" + key + "]")
	}

	return desc, nil
}

func (m DescriptionMap) GetOrNew(typ reflect.Type) (desc *Description) {
	key := typ.Name()
	desc, err := m.GetBy(key)
	if err == nil {
		return
	}

	desc = m.CreateByTyp(typ)
	return
}

func (m DescriptionMap) NewBy(key string) (interface{}, error) {
	desc, ok := m[key]
	if ok == false {
		return nil, errors.New("not found[" + key + "]")
	}

	return reflect.New(desc.Typ), nil
}
