package stcopy

import (
	"code.zhongguo168a.top/zg168a/gocodes/utils/convertutil"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
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
						x = reflect.ValueOf(convertutil.Convert2String(k.Interface()))
					case reflect.Int:
						x = reflect.ValueOf(convertutil.Convert2Int64(k.Interface()))

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

//
//func (ctx *Context) copy(source, target reflect.Value, provideTyp reflect.Type, depth int) (result reflect.Value, err error) {
//	fmt.Println("||| copy", "provide=", provideTyp)
//	fmt.Println("srctyp=", source.Type(), "src=", source)
//	fmt.Println("tartyp=", (Value)(target).GetTypeString(), "tar=", target, ",  canset=", target.CanSet())
//
//	if source.IsValid() == false || isValueNil(source) {
//		return
//	}
//
//	// 如果源与目标的类型不一致
//	// 0层不可以convert, 直接调用Convert函数处理
//	if depth != 0 && source.Kind() != target.Kind() {
//		switch ctx.direction {
//		case AtoB:
//			_, ok := source.Type().MethodByName("ConvertTo")
//			if ok == true {
//				methodVal := source.MethodByName("ConvertTo")
//				results := methodVal.Call([]reflect.Value{reflect.ValueOf(ctx.Params)})
//				result = results[0]
//				return
//			}
//		case AfromB:
//			_, ok := source.Type().MethodByName("ConvertFrom")
//			if ok == true {
//				methodVal := source.MethodByName("ConvertFrom")
//				results := methodVal.Call([]reflect.Value{target, reflect.ValueOf(ctx.Params)})
//				result = results[0]
//				return
//			}
//		}
//	}
//
//	// 默认
//	//
//	switch provideTyp.Kind() {
//	case reflect.Interface:
//		if target.IsValid() == false || target.CanSet() == false {
//			target = reflect.New(provideTyp).Elem()
//		}
//	case reflect.Map:
//	case reflect.Struct:
//		if target.IsValid() == false || target.IsNil() {
//			target = reflect.New(provideTyp).Elem()
//		}
//	case reflect.Ptr:
//		if target.IsValid() == false || target.IsNil() {
//			target = reflect.New(provideTyp).Elem()
//		}
//	default:
//		if target.IsValid() == false || target.CanSet() == false {
//			target = reflect.New(provideTyp).Elem()
//		}
//	}
//
//	if provideTyp.Kind() != reflect.Interface {
//		if source.Kind() == reflect.Interface {
//			source = source.Elem()
//		}
//		if target.Kind() == reflect.Interface {
//			target = target.Elem()
//		}
//	}
//
//	var retval reflect.Value
//	switch provideTyp.Kind() {
//	case reflect.Slice, reflect.Array:
//		if source.Len() == 0 {
//			return
//		}
//		for i := 0; i < source.Len(); i++ {
//			srcitem := source.Index(i)
//			taritem := func() (x reflect.Value) {
//				x = target.Index(i)
//				if x.IsValid() == false || x.IsNil() || x.CanSet() == false {
//					x = newMapValue(provideTyp.Elem(), srcitem, false)
//				}
//				return
//			}()
//			ctx.copy(srcitem, taritem, provideTyp.Elem(), depth+1)
//			target.Index(i).Set(convertToMapValue(taritem))
//		}
//	case reflect.Interface:
//		//srcelem := source.Elem()
//		//tarelem := func() (x reflect.Value) {
//		//	x = target.Elem()
//		//	switch srcelem.Kind() {
//		//	case reflect.Map, reflect.Ptr, reflect.Interface:
//		//		if x.IsValid() == false || x.IsNil() {
//		//			x = newMapValue(srcelem.Type(), srcelem, true)
//		//		}
//		//	default:
//		//		x = newMapValue(srcelem.Type(), srcelem, true)
//		//	}
//		//	return
//		//}()
//		retval, err = ctx.copy(source.Elem(), target.Elem(), source.Elem().Type(), depth+1)
//		if err != nil {
//			return
//		}
//		target.Set(retval)
//	case reflect.Ptr:
//		retval, err = ctx.copy(source.Elem(), target.Elem(), provideTyp.Elem(), depth+1)
//		if err != nil {
//			return
//		}
//		target.Set(retval)
//	case reflect.Struct:
//		for i, n := 0, provideTyp.NumField(); i < n; i++ {
//			field := provideTyp.Field(i)
//			key := reflect.ValueOf(field.Name)
//			srcfield := getFieldVal(source, field)
//			if source.Kind() == reflect.Map {
//				if srcfield.IsValid() == false || srcfield.IsNil() {
//					continue
//				}
//			}
//			fmt.Println("||| copy struct field: ", field.Name, ", fieldtyp=", field.Type)
//			fmt.Println("src=", srcfield, ", typ=", srcfield.Type())
//
//			// 获取目标值
//			tarfield := getFieldVal(target, field)
//			fmt.Println("tar=", tarfield, ", typ=", )
//			new, _ := ctx.copy(srcfield, tarfield, field.Type, depth+1)
//
//			switch target.Kind() {
//			case reflect.Struct:
//				target.FieldByName(field.Name).Set(new)
//			case reflect.Map:
//				tarfield = convertToMapValue(new)
//				target.SetMapIndex(key, new)
//			}
//
//		}
//		return
//	case reflect.Map:
//		for _, k := range source.MapKeys() {
//
//			val1 := source.MapIndex(k)
//			if val1.IsValid() == false {
//				continue
//			}
//
//			val2 := func() (x reflect.Value) {
//				x = target.MapIndex(k)
//				if !x.IsValid() || x.IsNil() || x.CanSet() == false {
//					x = newMapValue(val1.Type(), val1, false)
//				}
//				return
//			}()
//
//			ctx.copy(val1, val2, val1.Type(), depth+1)
//			val2 = convertToMapValue(val2)
//			key := func() (x reflect.Value) {
//				if k.Type() != target.Type().Key() {
//					switch target.Type().Key().Kind() {
//					case reflect.String:
//						x = reflect.ValueOf(convertutil.Convert2String(k.Interface()))
//					case reflect.Int:
//						x = reflect.ValueOf(convertutil.Convert2Int64(k.Interface()))
//
//					}
//				} else {
//					x = k
//				}
//				return
//			}()
//			target.SetMapIndex(key, val2)
//		}
//
//	case reflect.Func:
//		panic("not suppor")
//	default:
//		target.Set(func() (x reflect.Value) {
//			// 规定的类型跟源类型不一致的情况
//			if source.Type() != provideTyp {
//				switch source.Type().Kind() {
//				case reflect.Interface:
//					x = source.Elem().Convert(provideTyp)
//				default:
//					x = source.Convert(provideTyp)
//				}
//			} else {
//				x = source
//			}
//			return
//		}())
//
//		fmt.Println("default set >", source, target, target.Type())
//	}
//
//	result = target.convertToMapValue(target)
//	return
//}
