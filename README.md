# stcopy
拷贝一个对象至另一个对象, 对象与对象的类型可以不一致


#### 概念

* json map: 通过标准库json解析生成的map. 
* ValueA/ValueB stcopy.New(ValueA).To/From(ValueB)


#### 类型对应的关系如下:

| Golang        |   json map|
|:-------------:| -----:|
| map| map[string]interface{} |
| slice/array      |   []interface{} |
| 所有数字类型      |    float64 |
| string         |    string |
| bool        |    bool |
| []byte      |    base64 string |
| struct        |  map[string]interface{}|
| interface{}(基础类型)   |  interface{}(json map对应类型)|
| interface{}(struct类型)    |  map[string]interface{"_type":"type name", "_ptr":boolean}


#### 已实现功能

* 拷贝或者覆盖目标相同属性的值, 目标未覆盖的属性值会保留. 可通过新建目标对象实现完整拷贝
* 对于类型不一致的情况, 可以在ValueA增加ToType()和FromType()方法实现转换 
否则默认使用reflect.Convert方法来转化(可能会导致异常) 
* 提供了valid()方法, 递归遍历所有属性的valid()方法(如果存在).  
  	  
#### 应用场景:

* copy map to json map
* copy struct to json map
* copy struct to struct 


#### 例子

可参考单元测试