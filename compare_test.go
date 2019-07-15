package stcopy

import (
	"fmt"
	"strconv"
	"testing"
)

func TestContext_Compare(t *testing.T) {
	sources := []interface{}{
		//
		&map[string]interface{}{"Int": 10, "Map": map[string]interface{}{"String": "test compare"}},
		&map[string]interface{}{"Int": 11},
		[]string{
			"/Int: not equal: 10 !=11(s/t)",
			"/Map: valid not match: true !=false(s/t)",
		},
		//
		&map[string]interface{}{"Map": map[string]interface{}{"String": "test compare"}},
		&map[string]interface{}{"Map": map[string]interface{}{"String": "test compare 2"}},
		[]string{
			"/Map/String: not equal: test compare !=test compare 2(s/t)",
		},
		//
		&map[string]interface{}{"Array": []interface{}{"String", "test compare"}},
		&map[string]interface{}{"Array": []interface{}{"String", "test compare 2"}},
		[]string{
			"/Array/1: not equal: test compare !=test compare 2(s/t)",
		},
		//
		&Struct{String: "test"},
		&Struct{String: "test2"},
		[]string{
			"/String: not equal: test !=test2(s/t)",
		},
	}

	for i := 0; i < len(sources); i += 3 {
		source := sources[i]
		result := sources[i+1]
		errsResult := sources[i+2].([]string)

		errs := New(source).Compare(result)
		if len(errsResult) != len(errs) {
			panic("not equal " + strconv.Itoa(i))
		}
		for i, val := range errsResult {
			fmt.Println("compare_test[25]>", i, val)
			if val != errs[i].Error() {
				panic("not equal " + strconv.Itoa(i))
			}
		}
	}
}
