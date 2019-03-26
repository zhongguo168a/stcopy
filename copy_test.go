package stcopy

import (
	"code.zhongguo168a.top/zg168a/gocodes/utils/debugutil"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

var types = NewTypeMap([]reflect.Type{
	reflect.TypeOf(Struct{}),
})

type TestEnum int8

const (
	TestEnum_A1 TestEnum = iota + 1
	TestEnum_A2
)

func (s TestEnum) ToString(ctx *Context) (r string) {
	switch s {
	case TestEnum_A1:
		return "A1"
	case TestEnum_A2:
		return "A2"
	}
	return
}

func (s TestEnum) FromString(ctx *Context, val string) (e TestEnum, err error) {
	switch val {
	case "A1":
		return TestEnum_A1, nil
	case "A2":
		return TestEnum_A2, nil
	}
	return
}

type ClassAny struct {
	Any interface{}
}

type ClassStruct struct {
	Struct *Struct
}

type ClassBase struct {
	Int    int
	String string
	Bool   bool
	Bytes  []byte
}

type ClassCombination struct {
	ClassBase
	ClassStruct
}

type ClassArray struct {
	Array       []string
	ArrayStruct []*Struct
}

type ClassConvert struct {
	Convert       *Convert
	ConvertDefine ConvertDefine
	Enum          TestEnum
}

type ClassConvert2 struct {
	Convert string
	Enum    int
}

type ClassMap struct {
	Map       map[string]string
	MapStruct map[string]*Struct
}

type ClassEnum struct {
	Enum    TestEnum
	EnumMap *map[TestEnum]bool
}

type Struct struct {
	String string
	Int    int
	Bool   bool
}

type Convert struct {
	Int int16
}

func (s *Convert) ToString(ctx *Context) (r string) {
	b, _ := json.Marshal(s)
	r = string(b)
	return
}

func (s *Convert) FromString(ctx *Context, val string) (err error) {
	err = json.Unmarshal([]byte(val), &s)
	if err != nil {
		return
	}
	return
}

type ConvertMap struct {
	Int int16
}

func (s *ConvertMap) ToMap(ctx *Context) (r map[string]interface{}) {
	return
}

func (s *ConvertMap) FromMap(ctx *Context, val map[string]interface{}) (err error) {
	return
}

type ConvertDefine int

func (s ConvertDefine) ToString() (r string) {
	return strconv.Itoa(int(s))
}

func (s ConvertDefine) FromString(val string) (r ConvertDefine, err error) {
	t, err := strconv.Atoi(val)
	if err != nil {
		return
	}
	r = ConvertDefine(t)
	return
}

func TestCopyMapToStructEnum(t *testing.T) {
	sources := []interface{}{
		// next
		&map[string]interface{}{"Enum": "A1"}, // source
		&ClassEnum{},                          // target
		&ClassEnum{Enum: TestEnum_A1},         // result
		// next
		&map[string]interface{}{"EnumMap": map[string]interface{}{"A2": true}}, // source
		&ClassEnum{}, // target
		&ClassEnum{EnumMap: &map[TestEnum]bool{TestEnum_A2: true}}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		err := New(sources[i+1]).From(sources[i])
		if err != nil {
			t.Error("stcopy: " + err.Error())
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			t.Error("not equal")
		}
	}
}

func TestCopyMapToStructCombination(t *testing.T) {

	sources := []interface{}{
		// next
		&map[string]interface{}{"String": "test 1", "Int": 1, "Struct": &map[string]interface{}{"String": "test struct", "Bool": false, "Int": int(100)}}, // source
		&ClassCombination{}, // target
		&ClassCombination{ClassBase: ClassBase{String: "test 1", Int: 1}, ClassStruct: ClassStruct{Struct: &Struct{String: "test struct", Int: 100}}}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		origin, _ := NewContext(sources[i])
		err := origin.To(sources[i+1])
		if err != nil {
			t.Error("stcopy: " + err.Error())
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			t.Error("not equal")
		}
	}
}

func TestCopyStructToStructCombination(t *testing.T) {

	sources := []interface{}{
		&ClassCombination{ClassBase: ClassBase{String: "test 1", Int: 1}}, // source
		&ClassCombination{}, // target
		&ClassCombination{ClassBase: ClassBase{String: "test 1", Int: 1}}, // result
		// next
		//&Class2{String: &Convert{Int: 3}, Int: 2}, // target
		//&ClassBase{Int: 1},                      // target
		//&ClassBase{String: `{"Int":3}`, Int: 1}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		origin, _ := NewContext(sources[i])
		err := origin.To(sources[i+1])
		if err != nil {
			t.Error("stcopy: " + err.Error())
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			t.Error("not equal")
		}
	}
}

func TestCopyStructToStruct(t *testing.T) {

	sources := []interface{}{
		&ClassBase{String: "test 1", Int: 1}, // source
		&ClassBase{String: "test 2", Int: 2}, // target
		&ClassBase{String: "test 1", Int: 1}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		origin, _ := NewContext(sources[i])
		err := origin.To(sources[i+1])
		if err != nil {
			t.Error("stcopy: " + err.Error())
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			t.Error("not equal")
		}
	}
}

func TestCopyStructToStructConvert(t *testing.T) {

	sources := []interface{}{
		&ClassConvert{Convert: &Convert{Int: 100}, Enum: TestEnum(100)}, // source
		&ClassConvert2{Convert: `{"Int":99}`, Enum: 99},                 // target
		&ClassConvert2{Convert: `{"Int":100}`, Enum: 100},               // result
	}

	for i := 0; i < len(sources); i += 3 {
		origin, _ := NewContext(sources[i])
		err := origin.To(sources[i+1])
		if err != nil {
			t.Error("stcopy: " + err.Error())
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			t.Error("not equal")
		}
	}
}

func TestCopyMapToStructBase(t *testing.T) {

	sources := []interface{}{
		//&ClassBase{String: "test 1", Int: 1}, // source
		//&ClassBase{String: "test 2", Int: 2}, // target
		//&ClassBase{String: "test 1", Int: 1}, // result
		// next
		//&map[string]interface{}{"String": "test 1", "Int": 1}, // source
		//&ClassBase{String: "test 2", Int: 2},                  // target
		//&ClassBase{String: "test 1", Int: 1},                  // result
		//
		&map[string]interface{}{"Bytes": []byte("test")}, // source
		&ClassBase{},                      // target
		&ClassBase{Bytes: []byte("test")}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		origin, _ := NewContext(sources[i])
		err := origin.To(sources[i+1])
		if err != nil {
			t.Error("stcopy: " + err.Error())
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			t.Error("not equal")
		}
	}
}

func TestCopyStructFromMapConvert(t *testing.T) {
	sources := []interface{}{
		//&map[string]interface{}{"Convert": `{"Int":100}`},              // source
		//&ClassConvert{Convert: &Convert{Int: 99}, Enum: TestEnum(99)},  // target
		//&ClassConvert{Convert: &Convert{Int: 100}, Enum: TestEnum(99)}, // result
		// next
		//&map[string]interface{}{"Enum": 100}, // source
		//&ClassConvert{Enum: TestEnum(99)},    // target
		//&ClassConvert{Enum: TestEnum(100)},   // result
		// next
		&map[string]interface{}{"ConvertDefine": "100"},                      // source
		&ClassConvert{ConvertDefine: ConvertDefine(99), Enum: TestEnum(99)},  // target
		&ClassConvert{ConvertDefine: ConvertDefine(100), Enum: TestEnum(99)}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		err := New(sources[i+1]).From(sources[i])
		if err != nil {
			t.Error("stcopy: " + err.Error())
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			t.Error("not equal")
		}
	}
}

func TestCopyStructToMapConvert(t *testing.T) {

	sources := []interface{}{
		//
		//&ClassConvert{Convert: &Convert{Int: 100}},                        // source
		//&map[string]interface{}{},                                         // target
		//&map[string]interface{}{"Convert": `{"Int":100}`, "Enum": int(0)}, // result
		//
		&ClassBase{Bytes: []byte("test")},                                                     // source
		&map[string]interface{}{"String": "test 1"},                                           // target
		&map[string]interface{}{"Bytes": "dGVzdA==", "String": "", "Int": 0.0, "Bool": false}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		err := New(sources[i]).To(sources[i+1])
		if err != nil {
			t.Error("stcopy: " + err.Error())
		}
		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			t.Error("not equal")
		}
	}
}

func TestCopyStructToMapBase(t *testing.T) {

	sources := []interface{}{
		//&ClassStruct{&Struct{String: "test struct", Int: 100}}, // source
		//&map[string]interface{}{},                              // target
		//&map[string]interface{}{"Struct": map[string]interface{}{"String": "test struct", "Bool": false, "Int": 100.0, "_ptr": true}}, // result
		// next 结构没有赋值的字段, 也会覆盖掉目标字段
		&ClassBase{Bytes: []byte("test"), Int: 1},                                             // source
		&map[string]interface{}{"String": "test 2", "Int": 2.0},                               // target
		&map[string]interface{}{"Bytes": "dGVzdA==", "String": "", "Int": 1.0, "Bool": false}, // result
		// next
		//&ClassBase{String: "test 1", Int: 1},                                        // source
		//&map[string]interface{}{"Int": 2, "Bool": true},                             // target
		//&map[string]interface{}{"String": "test 1", "Int": int64(1), "Bool": false}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		err := New(sources[i]).To(sources[i+1])
		if err != nil {
			t.Error("stcopy: " + err.Error())
			return
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])

		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			t.Error("not equal")
		}
	}
}

// 根据provideType转化map
//func TestCopyMapToMapBase(t *testing.T) {
//
//	sources := []interface{}{
//		&map[string]interface{}{"String": "test 1", "Int": 1, "Bool": false}, // source
//		&map[string]interface{}{}, // target
//		&map[string]interface{}{"String": "test 1", "Int": int(1), "Bool": false}, // result
//		//next struct
//		//&map[string]interface{}{"Struct": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int(1)}}, // source
//		//&map[string]interface{}{"Struct": &map[string]interface{}{"String": "test struct", "Bool": false}},                   // target
//		//&map[string]interface{}{"Struct": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int(1)}}, // result
//	}
//
//	for i := 0; i < len(sources); i += 3 {
//		err := New(sources[i]).WithProvideTyp(reflect.TypeOf(&ClassBase{})).To(sources[i+1])
//		if err != nil {
//			t.Error("stcopy: " + err.Error())
//		}
//		debugutil.PrintJson("result=", sources[i+2])
//		debugutil.PrintJson("target=", sources[i+1])
//		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
//			t.Error("not equal")
//		}
//	}
//}

// 根据provideType转化map
//func TestCopyMapToMapArray(t *testing.T) {
//
//	sources := []interface{}{
//		//next
//		&map[string]interface{}{"Array": []string{"c", "d"}},      // source
//		&map[string]interface{}{},                                 // target
//		&map[string]interface{}{"Array": []interface{}{"c", "d"}}, // result
//		////next
//		&map[string]interface{}{"Array": []string{"c", "d"}},      // source
//		&map[string]interface{}{"Array": []interface{}{"a"}},      // target
//		&map[string]interface{}{"Array": []interface{}{"c", "d"}}, // result
//		//next array struct
//		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1"}}},           // source
//		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1", "Int": 1}}}, // target
//		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1"}}},           // result
//		//next array struct
//		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1"}, &map[string]interface{}{"String": "2"}}}, // source
//		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1", "Int": 1}}},                               // target
//		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1"}, &map[string]interface{}{"String": "2"}}}, // result
//	}
//
//	for i := 0; i < len(sources); i += 3 {
//		origin, err := NewContext(sources[i])
//		if err != nil {
//			panic(err)
//		}
//		origin.WithProvideTyp(reflect.TypeOf(&ClassArray{}))
//		err = origin.To(sources[i+1])
//		if err != nil {
//			panic(err)
//		}
//		debugutil.PrintJson("result=", sources[i+2])
//		debugutil.PrintJson("target=", sources[i+1])
//		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
//			t.Error("not equal")
//		}
//	}
//}

// 根据provideType转化map
//func TestCopyMapToMapMap(t *testing.T) {
//
//	sources := []interface{}{
//		// next map
//		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1"}}, // source
//		&map[string]interface{}{"Map": map[string]interface{}{"a": "a2"}}, // target
//		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1"}}, // result
//		//// next map
//		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1", "c": "c3"}},            // source
//		&map[string]interface{}{"Map": map[string]interface{}{"a": "a2", "b": "b2"}},            // target
//		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1", "b": "b2", "c": "c3"}}, // result
//		//// next map struct
//		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int(1)}}}, // source
//		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct", "Bool": false}}},                 // target
//		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int(1)}}}, // result
//		// next map struct
//		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int(1)}, "c": &map[string]interface{}{"String": "test struct 1", "Int": int(1)}}},                                                                       // source
//		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct", "Bool": false}, "b": &map[string]interface{}{"String": "test struct", "Bool": false}}},                                                                                         // target
//		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int(1)}, "b": &map[string]interface{}{"String": "test struct", "Bool": false}, "c": &map[string]interface{}{"String": "test struct 1", "Int": int(1)}}}, // result
//	}
//
//	for i := 0; i < len(sources); i += 3 {
//		origin, err := NewContext(sources[i])
//		if err != nil {
//			panic(err)
//		}
//		origin.WithProvideTyp(reflect.TypeOf(&ClassMap{}))
//		err = origin.To(sources[i+1])
//		if err != nil {
//			panic(err)
//		}
//		debugutil.PrintJson("result=", sources[i+2])
//		debugutil.PrintJson("target=", sources[i+1])
//		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
//			t.Error("not equal")
//		}
//	}
//}

func TestCopyAnyToJsonMap(t *testing.T) {
	sources := []interface{}{
		"test string type",
		"test string type",
		int(10),
		float64(10),
		int32(10),
		float64(10),
		uint(10),
		float64(10),
		// 5
		uint32(10),
		float64(10),
		true,
		true,
		map[string]interface{}{"String": "test struct"},
		map[string]interface{}{"String": "test struct"},
		&map[string]interface{}{"String": "test struct"},
		map[string]interface{}{"String": "test struct", "_ptr": true},
		map[string]interface{}{
			"a": "test a",
			"b": "test b",
		},
		map[string]interface{}{
			"a": "test a",
			"b": "test b",
		},
		// 10
		map[string]interface{}{
			"a": &map[string]interface{}{"String": "test map struct a"},
			"b": &map[string]interface{}{"String": "test map struct b"},
		},
		map[string]interface{}{
			"a": map[string]interface{}{"String": "test map struct a", "_ptr": true},
			"b": map[string]interface{}{"String": "test map struct b", "_ptr": true},
		},
		Struct{String: "test struct", Int: 100, Bool: true},
		map[string]interface{}{"String": "test struct", "Int": float64(100), "Bool": true, "_type": "Struct"},
		&Struct{String: "test struct", Int: 100, Bool: true},
		map[string]interface{}{"String": "test struct", "Int": float64(100), "Bool": true, "_type": "Struct", "_ptr": true},
		[]string{"1", "2"},
		[]interface{}{"1", "2"},
		[]byte{0, 1, 2},
		"AAEC",
		[]int{1, 2},
		[]interface{}{1.0, 2.0},
		[]map[string]interface{}{
			{"String": "test struct"},
			{"String": "test struct"},
		},
		[]interface{}{
			map[string]interface{}{"String": "test struct"},
			map[string]interface{}{"String": "test struct"},
		},
		[]*map[string]interface{}{
			{"String": "test struct"},
			{"String": "test struct"},
		},
		[]interface{}{
			map[string]interface{}{"String": "test struct", "_ptr": true},
			map[string]interface{}{"String": "test struct", "_ptr": true},
		},
		[]Struct{
			{String: "test struct", Int: 100, Bool: true},
			{String: "test struct 2", Int: 200, Bool: true},
		},
		[]interface{}{
			map[string]interface{}{"String": "test struct", "Int": 100.0, "Bool": true, "_type": "Struct"},
			map[string]interface{}{"String": "test struct 2", "Int": 200.0, "Bool": true, "_type": "Struct"},
		},
		[]*Struct{
			{String: "test struct", Int: 100, Bool: true},
			{String: "test struct 2", Int: 200, Bool: true},
		},
		[]interface{}{
			map[string]interface{}{"String": "test struct", "Int": 100.0, "Bool": true, "_type": "Struct", "_ptr": true},
			map[string]interface{}{"String": "test struct 2", "Int": 200.0, "Bool": true, "_type": "Struct", "_ptr": true},
		},
	}

	// map转换成json map
	for i := 0; i < len(sources); i += 2 {
		source := &map[string]interface{}{
			"Any": sources[i],
		}
		target := &map[string]interface{}{}
		origin, _ := NewContext(source)
		origin.WithProvideTyp(reflect.TypeOf(&ClassAny{}))
		err := origin.To(target)
		if err != nil {
			panic(err)
		}

		resultMap := sources[i+1]
		targetAny := (*target)["Any"]
		debugutil.PrintJson("result=", resultMap)
		debugutil.PrintJson("target=", targetAny)
		if reflect.DeepEqual(resultMap, targetAny) == false {
			t.Error("not equal")
		}
	}

	// struct转换成json map
	for i := 0; i < len(sources); i += 2 {
		source := &ClassAny{
			Any: sources[i],
		}
		target := &map[string]interface{}{}
		err := New(source).WithTypeMap(types).To(target)
		if err != nil {
			panic(err)
		}

		resultMap := sources[i+1]
		targetAny := (*target)["Any"]
		debugutil.PrintJson("result=", resultMap)
		debugutil.PrintJson("target=", targetAny)
		if reflect.DeepEqual(resultMap, targetAny) == false {
			t.Error("not equal")
		}
	}

}

func TestCopyJsonMapToStruct(t *testing.T) {
	sources := []interface{}{
		"test string type",
		"test string type",
		true,
		true,
		map[string]interface{}{
			"String": "test struct",
		},
		map[string]interface{}{
			"String": "test struct",
		},
		&map[string]interface{}{
			"String": "test struct",
		},
		map[string]interface{}{
			"String": "test struct",
			"_ptr":   true,
		},
		map[string]interface{}{
			"a": "test a",
			"b": "test b",
		},
		map[string]interface{}{
			"a": "test a",
			"b": "test b",
		},
		map[string]interface{}{
			"a": &map[string]interface{}{"String": "test map struct a"},
			"b": &map[string]interface{}{"String": "test map struct b"},
		},
		map[string]interface{}{
			"a": map[string]interface{}{"String": "test map struct a", "_ptr": true},
			"b": map[string]interface{}{"String": "test map struct b", "_ptr": true},
		},
		Struct{String: "test struct", Int: 100, Bool: true},
		map[string]interface{}{"String": "test struct", "Int": float64(100), "Bool": true, "_type": "Struct"},
		&Struct{String: "test struct", Int: 100, Bool: true},
		map[string]interface{}{"String": "test struct", "Int": float64(100), "Bool": true, "_type": "Struct", "_ptr": true},
		[]interface{}{"1", "2"},
		[]interface{}{"1", "2"},
		[]interface{}{
			map[string]interface{}{"String": "test struct"},
			map[string]interface{}{"String": "test struct"},
		},
		[]interface{}{
			map[string]interface{}{"String": "test struct"},
			map[string]interface{}{"String": "test struct"},
		},
		[]interface{}{
			&map[string]interface{}{"String": "test struct"},
			&map[string]interface{}{"String": "test struct"},
		},
		[]interface{}{
			map[string]interface{}{"String": "test struct", "_ptr": true},
			map[string]interface{}{"String": "test struct", "_ptr": true},
		},
		[]interface{}{
			Struct{String: "test struct", Int: 100, Bool: true},
			Struct{String: "test struct 2", Int: 200, Bool: true},
		},
		[]interface{}{
			map[string]interface{}{"String": "test struct", "Int": 100.0, "Bool": true, "_type": "Struct"},
			map[string]interface{}{"String": "test struct 2", "Int": 200.0, "Bool": true, "_type": "Struct"},
		},
		[]interface{}{
			&Struct{String: "test struct", Int: 100, Bool: true},
			&Struct{String: "test struct 2", Int: 200, Bool: true},
		},
		[]interface{}{
			map[string]interface{}{"String": "test struct", "Int": 100.0, "Bool": true, "_type": "Struct", "_ptr": true},
			map[string]interface{}{"String": "test struct 2", "Int": 200.0, "Bool": true, "_type": "Struct", "_ptr": true},
		},
	}

	// json map转换成struct
	for i := 0; i < len(sources); i += 2 {
		source := &map[string]interface{}{
			"Any": sources[i+1],
		}
		target := &ClassAny{}
		origin, _ := NewContext(target)
		err := origin.WithTypeMap(types).From(source)
		if err != nil {
			panic(err)
		}

		debugutil.PrintJson("result=", sources[i])
		debugutil.PrintJson("target=", target.Any)
		fmt.Printf("copy_test[679]> %T\n", sources[i])
		fmt.Printf("copy_test[679]> %T\n", target.Any)
		if reflect.DeepEqual(sources[i], target.Any) == false {
			t.Error("not equal")
		}
	}
}

func TestCopyMapToMap(t *testing.T) {
	sources := []interface{}{
		//// next map
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1"}}, // source
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a2"}}, // target
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1"}}, // result
		// next map
		&map[string]interface{}{"Map": &map[string]interface{}{"a": "a1"}},              // source
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a2", "_ptr": true}}, // target
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1", "_ptr": true}}, // result
		//// next map
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1", "c": "c3"}},            // source
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a2", "b": "b2"}},            // target
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1", "b": "b2", "c": "c3"}}, // result
		//// next map struct
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int(1)}}},           // source
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": map[string]interface{}{"String": "test struct", "Bool": false, "_ptr": true}}},              // target
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": 1.0, "_ptr": true}}}, // result
		// next map struct
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int(1)}, "c": &map[string]interface{}{"String": "test struct 1", "Int": int(1)}}},                                                                                          // source
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": map[string]interface{}{"String": "test struct", "Bool": false}, "b": map[string]interface{}{"String": "test struct", "Bool": false}}},                                                                                                              // target
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": 1.0, "_ptr": true}, "b": map[string]interface{}{"String": "test struct", "Bool": false}, "c": map[string]interface{}{"String": "test struct 1", "Int": 1.0, "_ptr": true}}}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		err := New(sources[i]).To(sources[i+1])
		if err != nil {
			t.Error("stcopy: " + err.Error())
			return
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			t.Error("not equal")
		}
	}
}
