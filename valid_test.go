package stcopy

import (
	"errors"
	"fmt"
	"testing"
)

type ValidClass struct {
	A *ValidA

	B TypeB

	ArrayB []TypeB

	MapB map[string]TypeB
}

type ValidA struct {
	PropInt int
}

func (a *ValidA) Valid() (err error) {
	if a.PropInt > 0 {
		err = errors.New("out of range")
		return
	}
	return
}

type TypeB int

func (a TypeB) Valid() bool {
	if a > 0 {
		return false
	}
	return true
}

func TestContext_Valid(t *testing.T) {
	sources := []interface{}{
		//&ValidClass{},
		//nil,
		//&ValidClass{A: &ValidA{PropInt: 100}},
		//"A: out of range",
		//&ValidClass{B: TypeB(100)},
		//"B: out of range",
		//&ValidClass{ArrayB: []TypeB{TypeB(0), TypeB(100)}},
		//"ArrayB: at 1: out of range",
		&ValidClass{MapB: map[string]TypeB{"1": TypeB(0), "2": TypeB(100)}},
		"MapB: at 2: not true",
	}

	for i := 0; i < len(sources); i += 2 {
		source := sources[i]
		result := sources[i+1]
		if result == nil {
			if New(source).Valid() != nil {
				panic("")
			}
		} else {
			if New(source).Valid() == nil {
				panic("")
			}
			fmt.Println("valid_test[43]>", New(source).Valid().Error())
			if New(source).Valid().Error() != result {
				panic("")
			}
		}
	}

}
