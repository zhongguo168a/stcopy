# stcopy
拷贝/覆盖一个对象至另一个对象, 对象与对象的类型可以不一致, 尽量保存结构的类型信息, 以便反向转换.



#### 应用场景:

* 基本用法
    * 深度拷贝/覆盖一个map/struct到另一个map/struct, 由于stcopy提供了预转换的方法, 可减少使用时, 每次都转换的麻烦
    * 通过map修改struct: 设置了键值, 会修改对应键值的struct
    * 验证结构字段是否合法
    * 不包含fmt和json, 可用于转换成js的场景
    
* 实用场景
    * 通讯协议的转换和验证
        * json协议先转换成对应的结构, 然后进行验证
        * grpc协议可转换成对应的结构, 然后进行验证
        * ...
    * 配置数据的转换       
        * 使用json/xml配置数据, 解析成golang map后, 转换成对应的结构
        * ...
    * 数据库数据的转换
        * 从数据库中获取数据后, 转换成对应的结构, 可进行验证
        * 编写orm的时候可以使用
        * ...
    * 各层之间对象的转换
        * Server/Dao等, 不同层次之间, 快捷的进行转换
        * 由于预转换的功能, 可减少对象的定义数量, 例如数据库中保存的是字符串json格式, 可通过转换功能, 直接转换成结构

#### 安装
go get github.com/zhongguo168a/stcopy

#### 概念

为了方便描述, 约定以下概念

* json map: 约定通过stcopy输出的map, 与encoding/json.Unmarshal方法生成的map类型一致
* ValueA/ValueB stcopy.New(ValueA).To/From(ValueB), 进行复制时的两个对象


#### 字段类型对应的关系如下:

| Golang        |   json map|
|:-------------:| -----:|
| 所有map| map[string]interface{} |
| 所有slice/array      |   []interface{} |
| 所有数字类型      |    float64 |
| string         |    string |
| bool        |    bool |
| []byte      |    base64 string |
| struct        |  map[string]interface{}|
| interface{}(基础类型)   |  interface{}(json map对应类型)|
| interface{}(struct类型)    |  map[string]interface{"_type":"type name", "_ptr":boolean}


#### 已实现功能

* 拷贝或者覆盖目标相同字段的值, 目标未覆盖的字段值会保留. 可通过新建目标对象实现完整拷贝
* 如果map中存在_type字段, 需要通过WithTypeMap()方法, 增加该结构的反射信息, 才能正确转换成对应的结构
    * 如果没有, 则拷贝一份map
* 可通过配置中的FieldTag参数, 修改字段的名字

```
  	// 例如在结构的Id字段 定义了bson:"_id", 转换后的struct.Id 变成 map["_id"]
    
```

* 对于类型不一致的情况
    * 可以在ValueA增加To(ctx)和From(ctx)方法实现转换
    * 可通过WithParams()方法, 在ctx中获取params, 实现不同场景的转换
    * 如果没有设置To/From, 默认使用reflect.Convert方法来转化
    * 遇到转换成目标类型, 返回错误
* 如果字段中存在不需要/无法拷贝的类型(例如time.Time), 可以通过设置BaseTypes, 像int类型一样, 直接赋值过去 
* 提供了valid()方法, 深度优先, 递归遍历所有字段的valid()方法(如果存在)


#### 例子

可参考单元测试, 单元测试集合了所有功能的示例