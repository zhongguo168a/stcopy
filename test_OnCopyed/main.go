package main

import (
	"github.com/zhongguo168a/stcopy"
)

type A struct {
	Int int // 初始值 = 10
}

func (a *A) OnCopyed() { // 首先进入
	a.Int = a.Int * 2 // 10 * 2 = 20
}

type B struct {
	Struct *A
}

func (b *B) OnCopyed() { // 最后进入
	b.Struct.Int = b.Struct.Int * 2 // 20 * 2 = 40
}

func main() {
	b := &B{}

	_ = stcopy.New(b).From(map[string]interface{}{
		"_ptr": true, "_type": "B",
		"Struct": map[string]interface{}{"_ptr": true, "_type": "A",
			"Int": 10}})
	println(b.Struct.Int) // echo 40
}
