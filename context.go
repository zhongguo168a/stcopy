package stcopy

import (
	"errors"
	"reflect"
)

//func NewContext(val interface{}) (ctx *Context, err error) {
//	ref := reflect.ValueOf(val)
//
//	if ref.Kind() != reflect.Ptr {
//		err = errors.New("origin must ptr struct or map")
//		return
//	}
//
//	unfold := TypeUtiler.UnfoldType(ref.Type())
//
//	if unfold.Kind() != reflect.Struct && unfold.Kind() != reflect.Map {
//		err = errors.New("origin must ptr struct or map")
//		return
//	}
//
//	ctx = &Context{
//		valueA: Value(ref),
//	}
//	return
//}

func New(val interface{}) (ctx *Context) {
	ref := reflect.ValueOf(val)
	return NewValue(ref)
}

func NewValue(ref reflect.Value) (ctx *Context) {
	if ref.Kind() != reflect.Ptr {
		panic(errors.New("origin must ptr struct or map"))
	}

	unfold := TypeUtiler.UnfoldType(ref.Type())
	if unfold.Kind() != reflect.Struct && unfold.Kind() != reflect.Map {
		panic(errors.New("origin must ptr struct or map"))
	}

	ctx = &Context{
		valueA:  Value(ref),
		Config:  NewConfig(),
		baseMap: NewTypeSet(),
	}
	return
}

type Direction int

const (
	AtoB Direction = iota
	AfromB
)

// 转换模式
type ConvertType int

const (
	AnyToJsonMap ConvertType = iota
	JsonMapToStruct
	StructToStruct
)

// 数据源的上下文
type Context struct {
	// 值A
	valueA Value
	// 值B
	valueB Value
	// copy方向
	direction Direction
	// 转换类型
	convertType ConvertType
	// 规定的类型
	provideTyp reflect.Type
	// 自定义的参数, 传递给转化函数使用
	params interface{}
	// 配置
	Config *Config
	// 类型的映射
	typeMap *TypeSet
	// 视作base类型的类型
	baseMap *TypeSet
}

func NewConfig() (obj *Config) {
	obj = &Config{}
	return
}

type Config struct {
	// 是否把枚举的值, 转成枚举的名字, 否则, 转化成枚举的值
	// 用于转化成配置文件的时候, 便于查阅
	// 需要使用 @description 标签的支持
	EnumToName bool
	// 转换成map时, 检查FieldTag定义的名字, 例如json/bson, 根据FieldTag转换成对应的Field名字
	// 例如在Id 字段 定义了bson:"_id", 转换后的map["Id"] 变成 map["_id"]
	FieldTag string
	// 当转化成map时, 是否总是携带结构信息, 包括_type和_ptr
	AlwaysStructInfo bool
	// 拷贝时, 如果来源的值等于默认值, 将被忽略
	// 可通过设置属性的tag: value:"" , 设置默认值. 如果没有设置, 根据属性类型的默认值
	IgnoreDefault bool
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

func (ctx *Context) GetParams() interface{} {
	return ctx.params
}

func (ctx *Context) WithProvideTyp(val reflect.Type) *Context {
	ctx.provideTyp = val
	return ctx
}

func (ctx *Context) WithTypeMap(val *TypeSet) *Context {
	ctx.typeMap = val
	return ctx
}

func (ctx *Context) WithBaseTypes(val *TypeSet) *Context {
	ctx.baseMap = val
	return ctx
}

func (ctx *Context) WithParams(val interface{}) *Context {
	ctx.params = val
	return ctx
}

func (ctx *Context) WithConfig(val *Config) *Context {
	ctx.Config = val
	return ctx
}

func (ctx *Context) WithFieldTag(tag string) *Context {
	ctx.Config.FieldTag = tag
	return ctx
}

func (ctx *Context) WithIgnoreDefault() *Context {
	ctx.Config.IgnoreDefault = true
	return ctx
}
