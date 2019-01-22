package stcopy

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
)

func isHard(k reflect.Kind) bool {
	switch k {
	case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Interface:
		return true
	}
	return false
}

func getFieldVal(val reflect.Value, field reflect.StructField) (x reflect.Value) {
	if val.Kind() == reflect.Map {
		x = val.MapIndex(reflect.ValueOf(field.Name))
	} else {
		x = val.FieldByName(field.Name)
	}
	return
}

func getTargetMode(val Value) TargetMode {
	if val.Upper().Kind() == reflect.Map {
		return TargetMap
	}
	return TargetStruct
}

func (ctx *Context) To(val interface{}) (err error) {
	ref := reflect.ValueOf(val)
	if ref.Kind() != reflect.Ptr {
		err = errors.New("target must ptr map")
		return
	}
	ctx.valueB = Value(ref)
	ctx.direction = AtoB

	if ctx.valueA.Indirect().Upper().Kind() == reflect.Map && ctx.valueB.Indirect().Upper().Kind() == reflect.Map {
		if ctx.provideTyp == nil {
			err = errors.New("must set provide type")
			return
		}
	} else {
		provideTyp, geterr := ctx.getProvideTyp(ctx.valueA, ctx.valueB)
		if geterr != nil {
			return
		}
		ctx.provideTyp = provideTyp
	}
	ctx.targetMode = getTargetMode(ctx.valueB.Indirect())

	_, err = ctx.copy(ctx.valueA, ctx.valueB, ctx.provideTyp, 0)
	if err != nil {
		return
	}
	return
}

func (ctx *Context) From(val interface{}) (err error) {
	ref := reflect.ValueOf(val)
	if ref.Kind() != reflect.Ptr {
		err = errors.New("target must ptr map")
		return
	}
	ctx.valueB = Value(ref)
	ctx.direction = AfromB
	if ctx.valueA.Indirect().Upper().Kind() == reflect.Map && ctx.valueB.Indirect().Upper().Kind() == reflect.Map {
		if ctx.provideTyp == nil {
			err = errors.New("must set provide type")
			return
		}
	}
	ctx.targetMode = getTargetMode(ctx.valueA.Indirect())
	provideTyp, geterr := ctx.getProvideTyp(ctx.valueB, ctx.valueA)
	if geterr != nil {
		return
	}
	_, err = ctx.copy(ctx.valueB, ctx.valueA, provideTyp, 0)
	if err != nil {
		return
	}
	return
}

func (ctx *Context) copy(source, target Value, provideTyp reflect.Type, depth int) (result Value, err error) {
	srcref := source.Upper()
	tarref := target.Upper()
	fmt.Println("\n||| to", "provide=", provideTyp)
	fmt.Println("srctyp=", srcref.Type(), "src=", srcref)
	fmt.Println("tartyp=", target.GetTypeString(), "tar=", tarref, ",  canset=", tarref.CanSet())

	// 源是否空
	if srcref.IsValid() == false {
		return
	}
	if isHard(srcref.Kind()) && srcref.IsNil() {
		return
	}

	// 接口处理
	if provideTyp.Kind() != reflect.Interface {
		if srcref.Kind() == reflect.Interface {
			srcref = srcref.Elem()
		}
		if tarref.Kind() == reflect.Interface {
			tarref = tarref.Elem()
		}
	}

	// 如果源与目标的类型不一致
	// 0层不可以convert, 直接调用Convert函数处理
	if depth != 0 && srcref.Kind() != tarref.Kind() {
		switch ctx.direction {
		case AtoB:
			_, ok := srcref.Type().MethodByName("ConvertTo")
			if ok == true {
				methodVal := srcref.MethodByName("ConvertTo")
				results := methodVal.Call([]reflect.Value{reflect.ValueOf(ctx)})
				result = Value(results[0])
				return
			}
		case AfromB:
			_, ok := tarref.Type().MethodByName("ConvertFrom")
			if ok == true {
				methodVal := tarref.MethodByName("ConvertFrom")
				results := methodVal.Call([]reflect.Value{reflect.ValueOf(ctx), srcref})
				if results[0].IsNil() == false {
					err = results[0].Interface().(error)
					return
				}
				result = Value(tarref)
				return
			}
		}
	}

	// 检查目标是否需要创建新的值
	checkNewTarget := func() (x bool) {
		if tarref.IsValid() == false {
			return true
		}
		switch provideTyp.Kind() {
		case reflect.Struct:

		case reflect.Map, reflect.Ptr:
			x = tarref.IsNil()
		case reflect.Interface:
			x = tarref.IsNil() || tarref.CanSet() == false
		default:
			x = tarref.CanSet() == false
		}

		return
	}()

	if checkNewTarget {
		// 创建新的值

		switch ctx.targetMode {
		case TargetMap:
			unfold := TypeSv.UnfoldType(provideTyp)
			switch unfold.Kind() {
			case reflect.Array, reflect.Slice:
				slice := make([]interface{}, srcref.Len(), srcref.Cap())
				tarref = reflect.ValueOf(&slice)
			case reflect.Map:
				tarref = reflect.ValueOf(&map[string]interface{}{})
			case reflect.Struct:
				tarref = reflect.ValueOf(&map[string]interface{}{})
			default:
				tarref = reflect.New(provideTyp)
			}

			if provideTyp.Kind() != reflect.Ptr {
				tarref = tarref.Elem()
			}
		case TargetStruct:
			tarref = reflect.New(provideTyp).Elem()
		}
	}
	fmt.Println("last target=", tarref, tarref.Type(), tarref.CanSet())

	var retval Value
	switch provideTyp.Kind() {
	case reflect.Slice, reflect.Array:
		if srcref.Len() == 0 {
			return
		}
		for i := 0; i < srcref.Len(); i++ {
			srcitem := srcref.Index(i)
			taritem := tarref.Index(i)
			retval, copyerr := ctx.copy(Value(srcitem), Value(taritem), provideTyp.Elem(), depth+1)
			if copyerr != nil {
				err = copyerr
				return
			}
			tarref.Index(i).Set(retval.Upper())
		}
	case reflect.Interface:
		retval, err = ctx.copy(Value(srcref.Elem()), Value(tarref.Elem()), srcref.Elem().Type(), depth+1)
		if err != nil {
			return
		}
		err = retval.updateMapStructTypeBy(Value(srcref.Elem()))
		if err != nil {
			return
		}
		tarref.Set(retval.Upper())
	case reflect.Ptr:
		retval, err = ctx.copy(Value(srcref.Elem()), Value(tarref.Elem()), provideTyp.Elem(), depth+1)
		if err != nil {
			return
		}
		if tarref.CanSet() {
			tarref.Set(retval.Upper().Addr())
		}
	case reflect.Struct:
		for i, n := 0, provideTyp.NumField(); i < n; i++ {
			field := provideTyp.Field(i)
			key := reflect.ValueOf(field.Name)
			srcfield := getFieldVal(srcref, field)
			if srcref.Kind() == reflect.Map {
				if srcfield.IsValid() == false || srcfield.IsNil() {
					continue
				}
			}
			fmt.Println(">>> copy struct field: ", field.Name, ", fieldtyp=", field.Type)
			// 获取目标值
			tarfield := getFieldVal(tarref, field)
			retval, err = ctx.copy(Value(srcfield), Value(tarfield), field.Type, depth+1)
			if err != nil {
				return
			}

			fmt.Println("copytomap[286]>", key, retval.Upper())
			switch tarref.Kind() {
			case reflect.Map:
				tarref.SetMapIndex(key, retval.Upper())
			case reflect.Struct:
				if retval.Upper().IsValid() {
					tarfield.Set(retval.Upper())
				}
			default:
				panic("not support")
			}
		}
	case reflect.Map:
		for _, k := range srcref.MapKeys() {

			val1 := srcref.MapIndex(k)
			if val1.IsValid() == false {
				continue
			}
			val2 := tarref.MapIndex(k)

			fmt.Println("||| copy map key: ", k, ", fieldtyp=", val1.Type())
			fmt.Println("src=", val1, ", typ=", val2)

			retval, _ := ctx.copy(Value(val1), Value(val2), val1.Type(), depth+1)
			key := func() (x reflect.Value) {
				if k.Type() != tarref.Type().Key() {
					switch tarref.Type().Key().Kind() {
					case reflect.String:
						x = reflect.ValueOf(Convert2String(k.Interface()))
					case reflect.Int:
						x = reflect.ValueOf(Convert2Int64(k.Interface()))

					}
				} else {
					x = k
				}
				return
			}()
			tarref.SetMapIndex(key, retval.Upper())
		}

	case reflect.Func:
		panic("not suppor")
	default:
		tarref.Set(func() (x reflect.Value) {
			// 规定的类型跟源类型不一致的情况
			if srcref.Type() != provideTyp {
				switch srcref.Type().Kind() {
				case reflect.Interface:
					x = srcref.Elem().Convert(provideTyp)
				default:
					x = srcref.Convert(provideTyp)
				}
			} else {
				x = srcref
			}
			return
		}())

	}

	result = Value(tarref)
	if ctx.targetMode == TargetMap {
		result = result.convertToMapValue()
	}

	fmt.Println("resut >", result.Upper())
	return
}
func Convert2String(ival interface{}) string {
	val := ""
	switch x := ival.(type) {
	case int:
		val = strconv.Itoa(x)
	case int8:
		val = strconv.Itoa(int(x))
	case int16:
		val = strconv.Itoa(int(x))
	case int32:
		val = strconv.Itoa(int(x))
	case int64:
		val = strconv.Itoa(int(x))
	case uint:
		val = strconv.Itoa(int(x))
	case uint8:
		val = strconv.Itoa(int(x))
	case uint16:
		val = strconv.Itoa(int(x))
	case uint32:
		val = strconv.Itoa(int(x))
	case uint64:
		val = strconv.Itoa(int(x))
	case float32:
		val = strconv.Itoa(int(x))
	case float64:
		val = strconv.Itoa(int(x))
	}
	return val
}

func Convert2Int64(ival interface{}) int64 {
	val := int64(0)
	switch data := ival.(type) {
	case int:
		val = int64(data)
	case int8:
		val = int64(data)
	case int16:
		val = int64(data)
	case int32:
		val = int64(data)
	case int64:
		val = data
	case uint:
		val = int64(data)
	case uint8:
		val = int64(data)
	case uint16:
		val = int64(data)
	case uint32:
		val = int64(data)
	case uint64:
		val = int64(data)
	case float32:
		val = int64(data)
	case float64:
		val = int64(data)
	case string:
		i, err := strconv.Atoi(data)
		if err != nil {
			panic(err)
		}
		val = int64(i)
	}

	return val
}
