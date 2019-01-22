package stcopy

import (
	"code.zhongguo168a.top/zg168a/gocodes/utils/debugutil"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type TestEnum int8

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
	Struct *Struct
}

type ClassArray struct {
	Array       []string
	ArrayStruct []*Struct
}

type ClassConvert struct {
	Convert *Convert
	Enum    TestEnum
}

type ClassConvert2 struct {
	Convert string
	Enum    int
}

type ClassMap struct {
	Map       map[string]string
	MapStruct map[string]*Struct
}

type Struct struct {
	String string
	Int    int
	Bool   bool
}

type Convert struct {
	Int int16
}

func (s *Convert) ConvertTo(ctx *Context) string {
	fmt.Println("copytomap_test[54]>")
	b, _ := json.Marshal(s)
	return string(b)
}

func (s *Convert) ConvertFrom(ctx *Context, val string) (err error) {
	err = json.Unmarshal([]byte(val), &s)
	if err != nil {
		return
	}
	return
}

func TestCopyStructToStruct(t *testing.T) {

	sources := []interface{}{
		&ClassBase{String: "test 1", Int: 1, Struct: &Struct{Int: 100, String: "test a", Bool: true}}, // source
		&ClassBase{String: "test 2", Int: 2}, // target
		&ClassBase{String: "test 1", Int: 1, Struct: &Struct{Int: 100, String: "test a", Bool: true}}, // result
		// next
		//&Class2{String: &Convert{Int: 3}, Int: 2}, // target
		//&ClassBase{Int: 1},                      // target
		//&ClassBase{String: `{"Int":3}`, Int: 1}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		origin, _ := NewContext(sources[i])
		err := origin.To(sources[i+1])
		if err != nil {
			panic(err)
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			panic("")
		}
	}
}

func TestCopyStructToStructConvert(t *testing.T) {

	sources := []interface{}{
		&ClassConvert{Convert: &Convert{Int: 100}, Enum: TestEnum(100)}, // source
		&ClassConvert2{Convert: `{"Int":99}`, Enum: 99},                 // target
		&ClassConvert2{Convert: `{"Int":100}`, Enum: 100},               // result
		// next
		//&Class2{String: &Convert{Int: 3}, Int: 2}, // target
		//&ClassBase{Int: 1},                      // target
		//&ClassBase{String: `{"Int":3}`, Int: 1}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		origin, _ := NewContext(sources[i])
		err := origin.To(sources[i+1])
		if err != nil {
			panic(err)
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			panic("")
		}
	}
}

func TestCopyMapToStructBase(t *testing.T) {

	sources := []interface{}{
		//&ClassBase{String: "test 1", Int: 1}, // source
		//&ClassBase{String: "test 2", Int: 2}, // target
		//&ClassBase{String: "test 1", Int: 1}, // result
		// next
		&map[string]interface{}{"String": "test 1", "Int": 1}, // source
		&ClassBase{String: "test 2", Int: 2},                  // target
		&ClassBase{String: "test 1", Int: 1},                  // result
	}

	for i := 0; i < len(sources); i += 3 {
		origin, _ := NewContext(sources[i])
		err := origin.To(sources[i+1])
		if err != nil {
			panic(err)
		}

		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			panic("")
		}
	}
}

func TestCopyStructFromMapConvert(t *testing.T) {
	sources := []interface{}{
		&map[string]interface{}{"Convert": `{"Int":100}`},              // source
		&ClassConvert{Convert: &Convert{Int: 99}, Enum: TestEnum(99)},  // target
		&ClassConvert{Convert: &Convert{Int: 100}, Enum: TestEnum(99)}, // result
		// next
		&map[string]interface{}{"Enum": 100}, // source
		&ClassConvert{Enum: TestEnum(99)},    // target
		&ClassConvert{Enum: TestEnum(100)},   // result
	}

	for i := 0; i < len(sources); i += 3 {
		origin, err := NewContext(sources[i+1])
		if err != nil {
			panic(err)
		}
		err = origin.From(sources[i])
		if err != nil {
			panic(err)
		}
		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			panic("")
		}
	}
}

func TestCopyStructToMapConvert(t *testing.T) {

	sources := []interface{}{
		&ClassConvert{Convert: &Convert{Int: 100}},                          // source
		&map[string]interface{}{},                                           // target
		&map[string]interface{}{"Convert": `{"Int":100}`, "Enum": int64(0)}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		err := New(sources[i]).To(sources[i+1])
		if err != nil {
			panic(err)
		}
		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			panic("")
		}
	}
}

func TestCopyStructToMapBase(t *testing.T) {

	sources := []interface{}{
		&ClassStruct{&Struct{String: "test struct", Int: 100}}, // source
		&map[string]interface{}{},                              // target
		&map[string]interface{}{"Struct": &map[string]interface{}{"String": "test struct", "Bool": false, "Int": int64(100)}}, // result
		// next
		&ClassBase{String: "test 1", Int: 1, Struct: &Struct{String: "test struct", Int: 100}}, // source
		&map[string]interface{}{}, // target
		&map[string]interface{}{"String": "test 1", "Int": int64(1), "Bool": false, "Struct": &map[string]interface{}{"String": "test struct", "Bool": false, "Int": int64(100)}}, // result
		// next
		&ClassBase{String: "test 1", Int: 1, Struct: &Struct{String: "test struct", Int: 100}},                                                                                    // source
		&map[string]interface{}{"Int": 2, "Bool": true, "Struct": &map[string]interface{}{"String": "old", "Bool": true, "Int": int64(0)}},                                        // target
		&map[string]interface{}{"String": "test 1", "Int": int64(1), "Bool": false, "Struct": &map[string]interface{}{"String": "test struct", "Bool": false, "Int": int64(100)}}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		err := New(sources[i]).To(sources[i+1])
		if err != nil {
			panic(err)
		}
		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			panic("")
		}
	}
}

func TestCopyMapToMapBase(t *testing.T) {

	sources := []interface{}{
		&map[string]interface{}{"String": "test 1", "Int": 1, "Bool": false, "Struct": &map[string]interface{}{"String": "test struct"}}, // source
		&map[string]interface{}{}, // target
		&map[string]interface{}{"String": "test 1", "Int": int64(1), "Bool": false, "Struct": &map[string]interface{}{"String": "test struct"}}, // result
		//next struct
		//&map[string]interface{}{"Struct": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int64(1)}}, // source
		//&map[string]interface{}{"Struct": &map[string]interface{}{"String": "test struct", "Bool": false}},                   // target
		//&map[string]interface{}{"Struct": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int64(1)}}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		err := New(sources[i]).WithProvideTyp(reflect.TypeOf(&ClassBase{})).To(sources[i+1])
		if err != nil {
			panic(err)
		}
		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			panic("")
		}
	}
}

func TestCopyMapToMapArray(t *testing.T) {

	sources := []interface{}{
		//next
		&map[string]interface{}{"Array": []string{"c", "d"}},      // source
		&map[string]interface{}{},                                 // target
		&map[string]interface{}{"Array": []interface{}{"c", "d"}}, // result
		////next
		&map[string]interface{}{"Array": []string{"c", "d"}},      // source
		&map[string]interface{}{"Array": []interface{}{"a"}},      // target
		&map[string]interface{}{"Array": []interface{}{"c", "d"}}, // result
		//next array struct
		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1"}}},           // source
		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1", "Int": 1}}}, // target
		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1"}}},           // result
		//next array struct
		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1"}, &map[string]interface{}{"String": "2"}}}, // source
		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1", "Int": 1}}},                               // target
		&map[string]interface{}{"ArrayStruct": []interface{}{&map[string]interface{}{"String": "1"}, &map[string]interface{}{"String": "2"}}}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		origin, err := NewContext(sources[i])
		if err != nil {
			panic(err)
		}
		origin.WithProvideTyp(reflect.TypeOf(&ClassArray{}))
		err = origin.To(sources[i+1])
		if err != nil {
			panic(err)
		}
		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			panic("")
		}
	}
}

func TestCopyMapToMapMap(t *testing.T) {

	sources := []interface{}{
		// next map
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1"}}, // source
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a2"}}, // target
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1"}}, // result
		//// next map
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1", "c": "c3"}},            // source
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a2", "b": "b2"}},            // target
		&map[string]interface{}{"Map": map[string]interface{}{"a": "a1", "b": "b2", "c": "c3"}}, // result
		//// next map struct
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int64(1)}}}, // source
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct", "Bool": false}}},                   // target
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int64(1)}}}, // result
		// next map struct
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int64(1)}, "c": &map[string]interface{}{"String": "test struct 1", "Int": int64(1)}}},                                                                       // source
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct", "Bool": false}, "b": &map[string]interface{}{"String": "test struct", "Bool": false}}},                                                                                             // target
		&map[string]interface{}{"MapStruct": map[string]interface{}{"a": &map[string]interface{}{"String": "test struct 1", "Bool": true, "Int": int64(1)}, "b": &map[string]interface{}{"String": "test struct", "Bool": false}, "c": &map[string]interface{}{"String": "test struct 1", "Int": int64(1)}}}, // result
	}

	for i := 0; i < len(sources); i += 3 {
		origin, err := NewContext(sources[i])
		if err != nil {
			panic(err)
		}
		origin.WithProvideTyp(reflect.TypeOf(&ClassMap{}))
		err = origin.To(sources[i+1])
		if err != nil {
			panic(err)
		}
		debugutil.PrintJson("result=", sources[i+2])
		debugutil.PrintJson("target=", sources[i+1])
		if reflect.DeepEqual(sources[i+2], sources[i+1]) == false {
			panic("")
		}
	}
}

func TestCopyStructAnyToMap(t *testing.T) {

	sources := []interface{}{
		//"test string type",
		//"test string type",
		//int(10),
		//int64(10),
		//int32(10),
		//int64(10),
		//uint(10),
		//uint64(10),
		//uint32(10),
		//uint64(10),
		//true,
		//true,
		//TestEnum(8),
		//int64(8),
		//&Struct{
		//	String: "test struct",
		//},
		//&map[string]interface{}{"_type": "Struct", "String": "test struct", "Bool": false, "Int": int64(0)},
		Struct{
			String: "test struct",
		},
		map[string]interface{}{"_type": "Struct", "String": "test struct", "Int": int64(0), "Bool": false},
		[]string{"1", "2"},
		[]interface{}{"1", "2"},
		map[string]string{
			"a": "test a",
			"b": "test b",
		},
		map[string]interface{}{
			"a": "test a",
			"b": "test b",
		},
		map[string]*Struct{
			"a": {String: "test map struct a"},
			"b": {String: "test map struct b"},
		},
		map[string]interface{}{
			"a": &map[string]interface{}{"String": "test map struct a", "Int": int64(0), "Bool": false},
			"b": &map[string]interface{}{"String": "test map struct b", "Int": int64(0), "Bool": false},
		},
	}

	for i := 0; i < len(sources); i += 2 {
		source := &ClassAny{
			Any: sources[i],
		}
		target := &map[string]interface{}{}
		origin, _ := NewContext(source)
		err := origin.To(target)
		if err != nil {
			panic(err)
		}

		debugutil.PrintJson("result=", sources[i+1])
		debugutil.PrintJson("target=", (*target)["Any"])
		if reflect.DeepEqual(sources[i+1], (*target)["Any"]) == false {
			panic("")
		}
	}
}

func TestCopyMapAnyToMap(t *testing.T) {
	sources := []interface{}{
		"test string type",
		"test string type",
		int(10),
		int64(10),
		int32(10),
		int64(10),
		uint(10),
		uint64(10),
		uint32(10),
		uint64(10),
		true,
		true,
		&map[string]interface{}{
			"String": "test struct",
		},
		&map[string]interface{}{
			"String": "test struct",
		},
		map[string]interface{}{
			"String": "test struct",
		},
		map[string]interface{}{
			"String": "test struct",
		},
		[]string{"1", "2"},
		[]interface{}{"1", "2"},
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
			"a": &map[string]interface{}{"String": "test map struct a"},
			"b": &map[string]interface{}{"String": "test map struct b"},
		},
	}

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

		debugutil.PrintJson("result=", sources[i])
		debugutil.PrintJson("target=", (*target)["Any"])
		if reflect.DeepEqual(sources[i+1], (*target)["Any"]) == false {
			panic("")
		}
	}
}
