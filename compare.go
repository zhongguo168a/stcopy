package stcopy

//
//func (ctx *Context) Compare(val interface{}) bool {
//	return ctx.compare(ctx.valueA, Value(reflect.ValueOf(val)), 0)
//}
//
//func (ctx *Context) compare(source, target Value, depth int) (b bool) {
//	srcref := source.Upper()
//	tarref := target.Upper()
//	//fmt.Println("\n||| to", "provide=", provideTyp)
//	//fmt.Println("srctyp=", srcref.Type(), "src=", srcref)
//
//	if srcref.Type() != tarref.Type() {
//		return false
//	}
//
//	switch srcref.Type().Kind() {
//	case reflect.Slice, reflect.Array:
//		if srcref.Len() == 0 {
//			return
//		}
//		for i := 0; i < srcref.Len(); i++ {
//			srcitem := srcref.Index(i)
//			err = ctx.valid(Value(srcitem), provideTyp.Elem(), depth+1)
//			if err != nil {
//				err = errors.New("at " + strconv.Itoa(i) + ": " + err.Error())
//				return
//			}
//		}
//	case reflect.Interface:
//		err = ctx.valid(Value(srcref.Elem()), srcref.Elem().Type(), depth+1)
//		if err != nil {
//			return
//		}
//	case reflect.Ptr:
//		err = ctx.valid(Value(srcref.Elem()), provideTyp.Elem(), depth+1)
//		if err != nil {
//			return
//		}
//	case reflect.Struct:
//		for _, field := range TypeUtiler.GetFieldRecursion(provideTyp) {
//			srcfield := getFieldVal(srcref, field)
//			if srcref.Kind() == reflect.Map {
//				if srcfield.IsValid() == false || srcfield.IsNil() {
//					continue
//				}
//			}
//			//fmt.Println(">>> copy struct field: ", field.Name, ", fieldtyp=", field.Type)
//			err = ctx.valid(Value(srcfield), field.Type, depth+1)
//			if err != nil {
//				err = errors.New(field.Name + ": " + err.Error())
//				return
//			}
//		}
//	case reflect.Map:
//		for _, k := range srcref.MapKeys() {
//			val1 := srcref.MapIndex(k)
//			if val1.IsValid() == false {
//				continue
//			}
//			//fmt.Println("||| copy map key: ", k, ", fieldtyp=", val1.Type())
//			//fmt.Println("src=", val1, ", typ=", val2)
//
//			err = ctx.valid(Value(val1), val1.Type(), depth+1)
//			if err != nil {
//				err = errors.New("at " + k.String() + ": " + err.Error())
//				return
//			}
//		}
//
//	case reflect.Func:
//		panic("not suppor")
//	default:
//	}
//
//	//fmt.Println("resut >", result.Upper())
//	return
//}
