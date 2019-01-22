package stcopy

import (
	"errors"
	"reflect"
)

func NewContext(val interface{}) (ctx *Context, err error) {
	ref := reflect.ValueOf(val)

	if ref.Kind() != reflect.Ptr {
		err = errors.New("origin must ptr struct or map")
		return
	}

	unfold := TypeUtiler.UnfoldType(ref.Type())

	if unfold.Kind() != reflect.Struct && unfold.Kind() != reflect.Map {
		err = errors.New("origin must ptr struct or map")
		return
	}

	ctx = &Context{
		valueA: Value(ref),
	}
	return
}

func New(val interface{}) (ctx *Context) {
	ref := reflect.ValueOf(val)

	if ref.Kind() != reflect.Ptr {
		panic(errors.New("origin must ptr struct or map"))
	}

	unfold := TypeUtiler.UnfoldType(ref.Type())
	if unfold.Kind() != reflect.Struct && unfold.Kind() != reflect.Map {
		panic(errors.New("origin must ptr struct or map"))
	}

	ctx = &Context{
		valueA: Value(ref),
	}
	return
}

type Direction int

const (
	AtoB Direction = iota
	AfromB
)

type TargetMode int

const (
	TargetMap TargetMode = iota
	TargetStruct
)

// 数据源的上下文
type Context struct {
	// 值A
	valueA Value
	// 值B
	valueB Value
	// copy方向
	direction Direction
	//
	targetMode TargetMode
	// 使用指定的描述集合
	descMap *DescriptionMap
	// 规定的类型
	provideTyp reflect.Type
	// 自定义的参数, 传递给转化函数使用
	Params interface{}
	// 配置
	Config *Config
}

type Config struct {
	// 是否把枚举的值, 转成枚举的名字, 否则, 转化成枚举的值
	// 用于转化成配置文件的时候, 便于查阅
	// 需要使用 @description 标签的支持
	EnumToName bool
}

func (ctx *Context) getProvideTyp(src, tar Value) (typ reflect.Type, err error) {
	typ = ctx.provideTyp
	srcref := src.Upper()
	tarref := tar.Upper()
	if typ == nil {
		indirect := reflect.Indirect(tarref)
		if indirect.Kind() == reflect.Struct {
			typ = tarref.Type()
			return
		}

		indirect = reflect.Indirect(srcref)
		if indirect.Kind() == reflect.Struct {
			typ = srcref.Type()
			return
		}
	}

	err = errors.New("not found")
	return
}

func (ctx *Context) WithDescriptionMap(val *DescriptionMap) *Context {
	ctx.descMap = val
	return ctx
}

func (ctx *Context) WithProvideTyp(val reflect.Type) *Context {
	ctx.provideTyp = val
	return ctx
}

func (ctx *Context) WithParams(val interface{}) *Context {
	ctx.Params = reflect.ValueOf(val)
	return ctx
}

func (ctx *Context) WithConfig(val *Config) *Context {
	ctx.Config = val
	return ctx
}
