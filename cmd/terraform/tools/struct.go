package tools

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

type ResourceDataWrapper map[string]interface{}

func ListIndex0(values interface{}) interface{} {
	if values == nil {
		return nil
	} else if len(values.([]interface{})) == 0 {
		return nil
	} else {
		return values.([]interface{})[0]
	}
}

func SetWrapper(values interface{}) ResourceDataWrapper {
	if values == nil {
		return nil
	} else if values.(*schema.Set).Len() == 0 {
		return nil
	} else {
		return values.(*schema.Set).List()[0].(map[string]interface{})
	}
}

func ListWrapper(values interface{}) ResourceDataWrapper {
	out, match := ListIndex0(values).(map[string]interface{})
	if !match {
		return nil
	}
	return out
}

func (this *ResourceDataWrapper) Get(name string) interface{} {
	return (*this)[name]
}

func (this ResourceDataWrapper) String(name, def string) string {
	if value, has := this[name]; !has {
		return def
	} else if str, match := value.(string); match && str != "" {
		return str
	} else {
		return def
	}
}

func (this ResourceDataWrapper) Int(name string) int {
	if value, has := this[name]; !has {
		return 0
	} else if str, match := value.(int); match && str != 0 {
		return str
	} else {
		return 0
	}
}

func (this ResourceDataWrapper) Bool(name string) bool {
	if value, has := this[name]; !has {
		return false
	} else if str, match := value.(bool); match {
		return str
	} else {
		return false
	}
}

/*func (this ResourceDataWrapper) List(name string, def []string) []string {
	if value, has := this[name]; !has {
		return nil
	} else if str, match := value.([]interface{}); match {
		outs := make([]string, len(str))
		for i, s := range str {
			outs[i] = s.(string)
		}
		return outs
	} else {
		return nil
	}
}*/

func (this ResourceDataWrapper) Set(name string, def []string) []string {
	if value, has := this[name]; !has {
		return def
	} else if set, match := value.(*schema.Set); match {
		outs := make([]string, set.Len())
		for i, s := range set.List() {
			outs[i] = s.(string)
		}
		return outs
	} else {
		return def
	}
}
