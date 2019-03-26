package stcopy

import (
	"encoding/base64"
	"errors"
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

func getTargetMode(source, target Value) (r ConvertType) {
	if target.Upper().Kind() == reflect.Map {
		return AnyToJsonMap
	}
	if source.Upper().Kind() == reflect.Map && target.Upper().Kind() == reflect.Struct {
		return JsonMapToStruct
	}
	if source.Upper().Kind() == reflect.Struct && target.Upper().Kind() == reflect.Struct {
		return StructToStruct
	}
	panic("not support ConvertType: src kind=" + source.Upper().Kind().String() + ", tar kind=" + target.Upper().Kind().String())
	return
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
			ctx.provideTyp = ctx.valueA.Upper().Type()
			//err = errors.New("must set provide type")
			//return
		}
	} else {
		provideTyp, geterr := ctx.getProvideTyp(ctx.valueA, ctx.valueB)
		if geterr != nil {
			return
		}
		ctx.provideTyp = provideTyp
	}
	ctx.convertType = getTargetMode(ctx.valueA.Indirect(), ctx.valueB.Indirect())

	_, err = ctx.copy(ctx.valueA.Indirect(), ctx.valueB.Indirect().Indirect(), ctx.provideTyp.Elem(), false, 0)
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
			ctx.provideTyp = ctx.valueA.Upper().Type()
			//err = errors.New("must set provide type")
			//return
		}
	} else {
		ctx.provideTyp, err = ctx.getProvideTyp(ctx.valueB, ctx.valueA)
		if err != nil {
			err = errors.New("must set provide type")
			return
		}
	}
	ctx.convertType = getTargetMode(ctx.valueB.Indirect(), ctx.valueA.Indirect())

	_, err = ctx.copy(ctx.valueB.Indirect(), ctx.valueA.Indirect(), ctx.provideTyp.Elem(), false, 0)
	if err != nil {
		return
	}

	return
}

func (ctx *Context) getMethodType(val reflect.Value) (r string) {
	r = stringutil.UpperFirst(val.Kind().String())
	return
}

func (ctx *Context) copy(source, target Value, provideTyp reflect.Type, inInterface bool, depth int) (result Value, err error) {

	srcref := source.Upper()
	tarref := target.Upper()

	// 所有指针类型的map结构都转化成非指针类型
	//if provideTyp.Kind() == reflect.Ptr {
	//	provideTypElem := provideTyp.Elem()
	//	if provideTypElem.Kind() == reflect.Map {
	//		fmt.Println(prefix+"copy convert: provide typ: ", provideTyp, "->", provideTypElem)
	//		provideTyp = provideTypElem
	//		srcref = srcref.Elem()
	//	}
	//}

	//prefix := strings.Repeat("----", depth)
	//fmt.Println(prefix+"> copy:", "provide typ=", provideTyp, "kind=", provideTyp.Kind())
	//fmt.Println(prefix+"copy: srctyp=", srcref.Type(), "src=", srcref)
	//fmt.Println(prefix+"copy: tartyp=", target.GetTypeString(), "tar=", tarref, "nil=", ",  canset=", tarref.CanSet(), func() (x string) {
	//	if isHard(tarref.Kind()) && tarref.IsNil() {
	//		x = "isnil=true"
	//	} else {
	//		x = "isnil=false"
	//	}
	//	return
	//}())

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

	// 检查目标是否需要创建新的值
	checkNewTarget := func() (x bool) {
		if tarref.IsValid() == false {
			return true
		}
		switch tarref.Kind() {
		case reflect.Array, reflect.Slice:
			x = tarref.IsNil()
		case reflect.Struct:

		case reflect.Map, reflect.Ptr:
			x = tarref.IsNil()
		case reflect.Interface:
			if isHard(tarref.Elem().Kind()) {
				x = tarref.IsNil()
			} else {
				x = tarref.IsNil() || tarref.CanSet() == false
			}

		default:
			x = tarref.CanSet() == false
		}

		return
	}()

	if checkNewTarget {
		// 创建新的值
		unfold := TypeUtiler.UnfoldType(provideTyp)
		switch unfold.Kind() {
		case reflect.Array, reflect.Slice:
			tarref = func() (x reflect.Value) {
				switch ctx.convertType {
				case AnyToJsonMap:
					if provideTyp == bytesTyp {
						x = reflect.New(stringTyp).Elem()
					} else {
						slice := make([]interface{}, srcref.Len(), srcref.Cap())
						x = reflect.ValueOf(slice)
					}
				case StructToStruct:
					x = reflect.MakeSlice(provideTyp, srcref.Len(), srcref.Cap())
				case JsonMapToStruct:
					x = reflect.MakeSlice(provideTyp, srcref.Len(), srcref.Cap())
				}
				//if isHard(provideTyp.Elem().Kind()) {
				//	slice := make([]interface{}, srcref.Len(), srcref.Cap())
				//	x = reflect.ValueOf(slice)
				//} else {
				//	if srcref.Kind() == reflect.String {
				//		x = reflect.MakeSlice(provideTyp, 0, 0)
				//	} else {
				//		x = reflect.MakeSlice(provideTyp, srcref.Len(), srcref.Cap())
				//	}
				//}

				return
			}()
		case reflect.Map:
			switch ctx.convertType {
			case AnyToJsonMap:
				if provideTyp.Kind() == reflect.Ptr {
					tarref = reflect.ValueOf(&map[string]interface{}{})
				} else {
					tarref = reflect.ValueOf(map[string]interface{}{})
				}
			case StructToStruct:
				tarref = reflect.MakeMap(unfold)
			case JsonMapToStruct:
				if inInterface {
					tarref = func() (y reflect.Value) {
						if provideTyp.Kind() == reflect.Ptr {
							y = reflect.ValueOf(&map[string]interface{}{})
						} else {
							y = reflect.ValueOf(map[string]interface{}{})
						}
						return
					}()
				} else {
					tarref = func() (x reflect.Value) {
						if provideTyp.Kind() == reflect.Ptr {
							x = reflect.New(provideTyp.Elem()).Elem()
							a := reflect.MakeMap(unfold)
							x.Set(a)
							x = x.Addr()
						} else {
							x = reflect.MakeMap(unfold)
						}
						return
					}()

				}
			}
		case reflect.Struct:
			if ctx.convertType == AnyToJsonMap {
				if provideTyp.Kind() == reflect.Ptr {
					tarref = reflect.ValueOf(&map[string]interface{}{})
				} else {
					tarref = reflect.ValueOf(map[string]interface{}{})
				}
			} else {
				tarref = reflect.New(unfold)
				if provideTyp.Kind() != reflect.Ptr {
					tarref = tarref.Elem()
				}
			}
		default:
			tarref = reflect.New(provideTyp)
			if provideTyp.Kind() != reflect.Ptr {
				tarref = tarref.Elem()
			}
		}

		target = Value(tarref)
	}

	//fmt.Println(prefix+"copy: tartyp=", tarref.Type(), "tar=", tarref, ",  canset=", tarref.CanSet(), func() (x string) {
	//	if isHard(tarref.Kind()) && tarref.IsNil() {
	//		x = "isnil=true"
	//	} else {
	//		x = "isnil=false"
	//	}
	//	return
	//}(), "<last>")

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
						result = Value(results[0])

					}

					return
				}

			}
		}
	}

	var retval Value
	switch provideTyp.Kind() {
	case reflect.Slice, reflect.Array:
		if srcref.Len() == 0 {
			return
		}

		switch ctx.convertType {
		case AnyToJsonMap:
			if provideTyp == bytesTyp {
				tarref = reflect.ValueOf(base64.StdEncoding.EncodeToString(srcref.Interface().([]byte)))
			} else {
				for i := 0; i < srcref.Len(); i++ {
					srcitem := srcref.Index(i)
					taritem := tarref.Index(i)
					retval, copyerr := ctx.copy(Value(srcitem), Value(taritem), provideTyp.Elem(), inInterface, depth+1)
					if copyerr != nil {
						err = copyerr
						return
					}
					//err = retval.updateMapStructTypeBy(Value(srcitem))
					//if err != nil {
					//	return
					//}
					//if retval.Indirect().Upper().Kind() == reflect.Map {
					//	err = retval.updateMapStructTypeBy(Value(srcitem))
					//	if err != nil {
					//		return
					//	}
					//
					//	err = retval.updateMapPtrBy(Value(srcitem))
					//	if err != nil {
					//		return
					//	}
					//}
					//
					tarref.Index(i).Set(retval.Upper())
				}
			}
		case JsonMapToStruct:
			if isHard(provideTyp.Elem().Kind()) {
				for i := 0; i < srcref.Len(); i++ {
					srcitem := srcref.Index(i)
					taritem := tarref.Index(i)
					retval, copyerr := ctx.copy(Value(srcitem), Value(taritem), provideTyp.Elem(), inInterface, depth+1)
					if copyerr != nil {
						err = copyerr
						return
					}
					tarref.Index(i).Set(retval.Upper())
				}
			} else if provideTyp.Elem().Kind() == reflect.Uint8 && srcref.Type().Kind() == reflect.String {
				b, _ := base64.StdEncoding.DecodeString(srcref.String())
				tarref = reflect.ValueOf(b)
			} else {
				reflect.Copy(tarref, srcref)
			}
		}
	case reflect.Interface:
		switch ctx.convertType {
		case AnyToJsonMap:
			retval, err = ctx.copy(Value(srcref.Elem()), Value(tarref.Elem()), srcref.Elem().Type(), true, depth+1)
			if err != nil {
				return
			}
		case JsonMapToStruct:
			provideTyp = func() (x reflect.Type) {
				srcunfold := source.unfoldInterface()
				checkMap := func() (y bool) {
					switch srcunfold.Upper().Kind() {
					case reflect.Map:
						y = true
						return
					}
					return
				}()

				if checkMap {
					x, _ = source.parseMapType(ctx)
				} else {
					x = srcref.Elem().Type()
				}

				return
			}()

			retval, err = ctx.copy(Value(srcref.Elem()), Value(tarref.Elem()), provideTyp, true, depth+1)
			if err != nil {
				return
			}
		}
		//err = retval.updateMapStructPtrBy(Value(srcref.Elem()))
		//if err != nil {
		//	return
		//}
		if tarref.CanSet() {
			tarref.Set(retval.Upper())
		}
	case reflect.Ptr:
		srcptr := func() (x Value) {
			if srcref.Kind() == reflect.Ptr {
				x = Value(srcref.Elem())
			} else {
				x = Value(srcref)
			}
			return
		}()

		tarptr := func() (x Value) {
			if tarref.Kind() == reflect.Ptr {
				x = Value(tarref.Elem())
			} else {
				x = Value(tarref)
			}
			return
		}()
		retval, err = ctx.copy(srcptr, tarptr, provideTyp.Elem(), inInterface, depth+1)
		if err != nil {
			return
		}
		if tarref.Kind() == reflect.Ptr {
			tarptr.Upper().Set(retval.Upper())
		}

		switch ctx.convertType {
		case AnyToJsonMap:
			if retval.Indirect().Upper().Kind() == reflect.Map {
				err = retval.updateMapPtrBy(source.unfoldInterface())
				if err != nil {
					return
				}

				if tarref.Kind() == reflect.Ptr {
					tarref = tarref.Elem()
				}
			}
		}

		//if tarref.CanSet() {
		//	tarref.Set(retval.Upper().Addr())
		//}
	case reflect.Struct:
		for _, field := range TypeUtiler.GetFieldRecursion(provideTyp) {
			key := reflect.ValueOf(field.Name)
			srcfield := getFieldVal(srcref, field)
			if srcref.Kind() == reflect.Map {
				if srcfield.IsValid() == false || srcfield.IsNil() {
					continue
				}
			}
			//fmt.Println(prefix+"struct: field=", field.Name, ", fieldtyp=", field.Type)
			// 获取目标值
			tarfield := getFieldVal(tarref, field)
			retval, err = ctx.copy(Value(srcfield), Value(tarfield), field.Type, inInterface, depth+1)
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
				panic("not support in struct")
			}
			//if retval.Upper().IsValid() {
			//	fmt.Println(prefix+"struct[AF]: field=", field.Name, ", fieldval=", retval.Upper().Interface())
			//} else {
			//	fmt.Println(prefix+"struct[AF]: field=", field.Name, ", fieldval=", "not valid")
			//}
		}

		switch ctx.convertType {
		case AnyToJsonMap:
			if inInterface && tarref.Kind() == reflect.Map {
				err = Value(tarref).updateMapStructTypeBy(source.unfoldInterface())
				if err != nil {
					return
				}
			}
		}
	case reflect.Map:
		for _, keySrc := range srcref.MapKeys() {
			//fmt.Println(prefix+"map: before copy key source: type=", keySrc.Type(), ", val=", keySrc.Interface())
			keyTar := reflect.New(provideTyp.Key()).Elem()
			keyTarVal, copyerr := ctx.copy(Value(keySrc), Value(keyTar), provideTyp.Key(), inInterface, depth+1)
			if copyerr != nil {
				err = copyerr
				return
			}
			//fmt.Println(prefix+"map: after copy key target: type=", keyTarVal.Upper().Type(), ", val=", keyTarVal.Upper().Interface())

			valTar := tarref.MapIndex(keyTar)
			valSrc := srcref.MapIndex(keySrc)
			if valSrc.IsValid() == false {
				continue
			}

			//fmt.Println(prefix+"map: before copy value source : type=", valSrc.Type(), ", val=", valSrc.Interface())
			//if valTar.IsValid() && valTar.IsNil() == false {
			//	fmt.Println(prefix+"before copy value target: type=", valTar.Type(), ", val=", valTar.Interface())
			//}
			varTarVal, copyerr := ctx.copy(Value(valSrc), Value(valTar), provideTyp.Elem(), inInterface, depth+1)
			if copyerr != nil {
				err = copyerr
				return
			}
			//fmt.Println(prefix+"map: after copy value target: type=", varTarVal.Upper().Type(), ", val=", varTarVal.Upper().Interface())
			tarref.SetMapIndex(keyTarVal.Upper(), varTarVal.Upper())
			//fmt.Println(prefix+"map: after copy value map: ", tarref.Interface())
		}

	case reflect.Func:
		panic("function not support")
	default:
		tarref.Set(func() (x reflect.Value) {
			// 规定的类型跟源类型不一致的情况
			if srcref.Type() != provideTyp {
				switch srcref.Type().Kind() {
				case reflect.Interface:
					x = srcref.Elem().Convert(provideTyp)
				default:
					if provideTyp.Kind().String() == provideTyp.String() {
						x = srcref.Convert(provideTyp)
					} else {
						// 枚举
						err = errors.New("enum convert function not found")
						return
					}
				}
			} else {
				x = srcref
			}
			return
		}())

	}

	result = Value(tarref)
	if ctx.convertType == AnyToJsonMap {
		result = result.convertToMapValue()
	}

	//fmt.Println("resut >", result.Upper())
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
