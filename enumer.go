package stcopy

func NewEnum(args ...interface{}) *Enum {
	e := &Enum{
		Name2Value: map[string]interface{}{},
		Value2Name: map[interface{}]string{},
	}
	for i := 0; i < len(args); i += 2 {
		key := args[i].(string)
		val := args[i+1]
		e.Name2Value[key] = val
		e.Value2Name[val] = key
		e.names = append(e.names, key)
	}
	return e
}

type Enum struct {
	Name2Value map[string]interface{}
	Value2Name map[interface{}]string

	names []string
}

func (e *Enum) Names() (r []string) {
	return e.names
}

func (e *Enum) NameFirst() (r string) {
	return e.names[0]
}

func (e *Enum) ValueFirst() (r interface{}) {
	if len(e.names) > 0 {
		return e.Name2Value[e.names[0]]
	}

	return 0
}
