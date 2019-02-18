package stcopy

import (
	"github.com/pkg/errors"
	"github.com/zhongguo168a/gocodes/utils/stringutil"
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

func getFieldVal(val reflect.Value, field *reflect.StructField) (x reflect.Value) {
	if val.Kind() == reflect.Map {
		x = val.MapIndex(reflect.ValueOf(field.Name))
	} else {
		x = val.FieldByName(field.Name)
	}
	return
}

func getTargetMode(val Value) TargetType {
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
	ctx.targetType = getTargetMode(ctx.valueB.Indirect())

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
	} else {
		ctx.provideTyp, err = ctx.getProvideTyp(ctx.valueB, ctx.valueA)
		if err != nil {
			err = errors.New("must set provide type")
			return
		}
	}
	ctx.targetType = getTargetMode(ctx.valueA.Indirect())

	_, err = ctx.copy(ctx.valueB, ctx.valueA, ctx.provideTyp, 0)
	if err != nil {
		return
	}

	return
}

func (ctx *Context) copy(source, target Value, provideTyp reflect.Type, depth int) (result Value, err error) {
	srcref := source.Upper()
	tarref := target.Upper()
	//fmt.Println("\n||| to", "provide=", provideTyp)
	//fmt.Println("srctyp=", srcref.Type(), "src=", srcref)
	//fmt.Println("tartyp=", target.GetTypeString(), "tar=", tarref, ",  canset=", tarref.CanSet())

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
	// 0层不可以convert, 请直接调用Convert函数处理
	if depth != 0 && srcref.Kind() != tarref.Kind() {
		switch ctx.direction {
		case AtoB:
			mname := "To" + ctx.getMethodType(tarref)
			mtype, ok := srcref.Type().MethodByName(mname)
			if ok == true {
				methodVal := srcref.MethodByName(mname)
				if mtype.Type.NumIn() > 2 {
					err = errors.New("func " + mname + " NumIn() must 1 or 0")
					return
				}
				results := func() (x []reflect.Value) {
					if mtype.Type.NumIn() == 2 {
						x = methodVal.Call([]reflect.Value{reflect.ValueOf(ctx)})

					} else {
						x = methodVal.Call([]reflect.Value{})
					}
					return
				}()
				// 包含error
				if mtype.Type.NumOut() > 1 {
					if results[1].IsNil() == false {
						err = results[1].Interface().(error)
						return
					}
				}
				result = Value(results[0])
				return
			}
		case AfromB:
			if tarref.IsValid() == true {
				mname := "From" + ctx.getMethodType(srcref)
				mtype, ok := tarref.Type().MethodByName(mname)
				if ok == true {
					methodVal := tarref.MethodByName(mname)
					if mtype.Type.NumIn() > 3 || mtype.Type.NumIn() == 1 {
						err = errors.New("func " + mname + " NumIn() must 2 or 1")
						return
					}

					results := func() (x []reflect.Value) {
						if mtype.Type.NumIn() == 3 {
							x = methodVal.Call([]reflect.Value{reflect.ValueOf(ctx), srcref})
						} else {
							x = methodVal.Call([]reflect.Value{srcref})
						}
						return
					}()

					if mtype.Type.NumOut() > 0 {
						if tarref.Kind() == reflect.Ptr {
							if results[0].IsNil() == false {
								err = results[0].Interface().(error)
								return
							}
						} else {
							tarref.Set(results[0])
						}

					}

					return
				}

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

		switch ctx.targetType {
		case TargetMap:
			unfold := TypeUtiler.UnfoldType(provideTyp)
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
	//fmt.Println("last target=", tarref, tarref.Type(), tarref.CanSet())

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
		for _, field := range TypeUtiler.GetFieldRecursion(provideTyp) {
			key := reflect.ValueOf(field.Name)
			srcfield := getFieldVal(srcref, field)
			if srcref.Kind() == reflect.Map {
				if srcfield.IsValid() == false || srcfield.IsNil() {
					continue
				}
			}
			//fmt.Println(">>> copy struct field: ", field.Name, ", fieldtyp=", field.Type)
			// 获取目标值
			tarfield := getFieldVal(tarref, field)
			retval, err = ctx.copy(Value(srcfield), Value(tarfield), field.Type, depth+1)
			if err != nil {
				return
			}

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

			//fmt.Println("||| copy map key: ", k, ", fieldtyp=", val1.Type())
			//fmt.Println("src=", val1, ", typ=", val2)

			retval, _ := ctx.copy(Value(val1), Value(val2), val1.Type(), depth+1)
			key := func() (x reflect.Value) {
				if k.Type() != tarref.Type().Key() {
					switch tarref.Type().Key().Kind() {
					case reflect.String:
						x = reflect.ValueOf(Convert2String(k.Interface()))
					case reflect.Int:
						x = reflect.ValueOf(Convert2Int(k.Interface()))

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
	if ctx.targetType == TargetMap {
		result = result.convertToMapValue()
	}

	//fmt.Println("resut >", result.Upper())
	return
}

func (ctx *Context) getMethodType(val reflect.Value) (r string) {
	r = stringutil.UpperFirst(val.Kind().String())
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

func Convert2Int(ival interface{}) int {
	val := int(0)
	switch data := ival.(type) {
	case int:
		val = int(data)
	case int8:
		val = int(data)
	case int16:
		val = int(data)
	case int32:
		val = int(data)
	case int64:
		val = int(data)
	case uint:
		val = int(data)
	case uint8:
		val = int(data)
	case uint16:
		val = int(data)
	case uint32:
		val = int(data)
	case uint64:
		val = int(data)
	case float32:
		val = int(data)
	case float64:
		val = int(data)
	case string:
		i, err := strconv.Atoi(data)
		if err != nil {
			panic(err)
		}
		val = int(i)
	}

	return val
}
