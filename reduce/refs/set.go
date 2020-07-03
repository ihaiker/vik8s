package refs

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func reflectValue(obj interface{}) reflect.Value {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

func GetField(obj interface{}, name string) reflect.Value {
	objValue := reflectValue(obj)
	return objValue.FieldByName(name)
}

func tagField(obj interface{}, name string) (reflect.StructField, reflect.Value, bool) {
	objValue := reflectValue(obj)
	fieldsCount := objValue.Type().NumField()
	for i := 0; i < fieldsCount; i++ {
		field := objValue.Type().Field(i)
		tag := field.Tag.Get("json")
		fieldTagValue, _ := utils.Split2(tag, ",")
		if fieldTagValue == "-" || fieldTagValue == "" {
			if !isBase(field.Type) {
				v := objValue.Field(i)
				if field.Type.Kind() != reflect.Ptr {
					v = v.Addr()
				} else {
					v = reflect.New(field.Type.Elem())
				}
				sf, rv, has := tagField(v.Interface(), name)
				if !has {
					continue
				}
				if field.Type.Kind() == reflect.Ptr {
					objValue.Field(i).Set(v)
				}
				return sf, rv, has
			}
		} else if strings.ToLower(fieldTagValue) == strings.ToLower(name) {
			return field, objValue.Field(i), true
		}
	}
	return reflect.StructField{}, reflect.Value{}, false
}

//是否是简单类型
func isBase(fieldType reflect.Type) bool {
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}
	switch fieldType.Kind() {
	default:
		return false

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	case reflect.Float32, reflect.Float64:
	case reflect.Bool, reflect.String:
	case reflect.Struct:
		return fieldType.String() == "time.Time"
	}
	return true
}

func Unmarshal(obj interface{}, item *config.Directive) {
	UnmarshalWith(obj, item, Defaults)
}

func UnmarshalItem(obj interface{}, item *config.Directive) {
	UnmarshalItemWith(obj, item, Defaults)
}

func UnmarshalWith(obj interface{}, item *config.Directive, manager TypeManager) {
	for _, directive := range item.Body {
		UnmarshalItemWith(obj, directive, manager)
	}
}

func UnmarshalItemWith(obj interface{}, item *config.Directive, manager TypeManager) {
	utils.Assert(reflect.ValueOf(obj).Kind() == reflect.Ptr,
		"Object must must be interface: %v", reflect.TypeOf(obj))

	field, value, has := tagField(obj, item.Name)
	utils.Assert(has, "Invalid field：%s", item.Name)

	if v, has := manager.DealWith(field.Type, item); has {
		value.Set(v)
	} else if baseValue, err := baseValue(field.Type, utils.Index(item.Args, 0), manager); err == nil {
		value.Set(baseValue)
	} else if v, err := assemblyValue(field.Type, value, item, manager); err == nil {
		value.Set(v)
	} else {
		utils.Panic(err, "Invalid %s", item.Name)
	}
}

func assemblyValue(fieldType reflect.Type, value reflect.Value, item *config.Directive, manager TypeManager) (reflect.Value, error) {

	if fieldType.Kind() == reflect.Ptr {
		if out, err := assemblyValue(fieldType.Elem(), value.Elem(), item, manager); err == nil {
			v := reflect.New(fieldType.Elem())
			v.Elem().Set(out)
			return v, nil
		} else {
			return out, err
		}
	}

	if v, has := manager.DealWith(fieldType, item); has {
		return v, nil
	}

	switch fieldType.Kind() {
	case reflect.Array:
		//return reflect.Value{}, fmt.Errorf("Invalid %s", item.Name)

	case reflect.Map:
		v := reflect.New(fieldType)
		m := mapValue(fieldType.Key(), fieldType.Elem(), item, manager)
		if value.IsValid() {
			for mr := value.MapRange(); mr.Next(); {
				m.SetMapIndex(mr.Key(), mr.Value())
			}
		}
		v.Elem().Set(m)
		return v.Elem(), nil

	case reflect.Slice:
		v := reflect.New(fieldType)
		slice := sliceValue(fieldType.Elem(), item, manager)
		v.Elem().Set(slice)
		if value.IsValid() {
			v.Elem().Set(reflect.AppendSlice(value, slice))
		}
		return v.Elem(), nil

	case reflect.Struct:
		value := structValue(fieldType, item, manager)
		return value, nil
	}

	return reflect.Value{}, os.ErrInvalid
}

func structValue(fieldType reflect.Type, item *config.Directive, manager TypeManager) reflect.Value {
	value := reflect.New(fieldType)
	for _, arg := range item.Args {
		fieldName, fieldValue := utils.CompileSplit2(arg, ":|=")
		UnmarshalItemWith(value.Interface(), &config.Directive{
			Line: item.Line, Name: fieldName, Args: []string{fieldValue},
		}, manager)
	}
	for _, directive := range item.Body {
		UnmarshalItemWith(value.Interface(), directive, manager)
	}
	return value.Elem()
}

func mapValue(keyType, valueType reflect.Type, item *config.Directive, manager TypeManager) reflect.Value {
	m := reflect.MakeMap(reflect.MapOf(keyType, valueType))
	utils.Assert(isBase(keyType), "invalid %s, line %d", item.Name, item.Line)

	if isBase(valueType) {
		utils.Assert(len(item.Args) == 0 || (len(item.Args) == 2 && len(item.Body) == 0),
			"invalid %s, line %d", item.Name, item.Line)

		if len(item.Args) == 2 {
			key, err := baseValue(keyType, item.Args[0], manager)
			utils.Panic(err, "invalid %s %s, line %d", item.Name, strings.Join(item.Args, " "), item.Line)
			value, err := baseValue(valueType, item.Args[1], manager)
			utils.Panic(err, "invalid %s %s, line %d", item.Name, strings.Join(item.Args, " "), item.Line)
			m.SetMapIndex(key, value)
		} else {
			for _, d := range item.Body {
				utils.Assert(len(d.Args) == 1, "invalid %s %s, line %d", d.Name, strings.Join(d.Args, " "), d.Line)

				key, err := baseValue(keyType, d.Name, manager)
				utils.Panic(err, "invalid %s %s, line %d", d.Name, strings.Join(d.Args, " "), d.Line)

				value, err := baseValue(valueType, d.Args[0], manager)
				utils.Panic(err, "invalid %s %s, line %d", d.Name, strings.Join(d.Args, " "), d.Line)

				m.SetMapIndex(key, value)
			}
		}
	} else {
		if len(item.Args) == 1 {
			key, err := baseValue(keyType, item.Args[0], manager)
			utils.Panic(err, "invalid %s %s, line %d", item.Name, strings.Join(item.Args, " "), item.Line)

			vs := reflect.New(valueType)
			if vs.Kind() == reflect.Ptr {
				vs = vs.Elem()
			}
			item.Args = item.Args[1:]
			if value, has := manager.DealWith(valueType, item); has {
				m.SetMapIndex(key, value)
			} else {
				value, err := assemblyValue(valueType, vs, item, manager)
				utils.Panic(err, "invalid %s %s, line %d", item.Name, strings.Join(item.Args, " "), item.Line)
				m.SetMapIndex(key, value)
			}
		} else {
			for _, d := range item.Body {
				utils.Assert(len(d.Args) == 0, "invalid %s %s, line %d", d.Name, strings.Join(d.Args, " "), d.Line)

				key, err := baseValue(keyType, d.Name, manager)
				utils.Panic(err, "invalid %s %s, line %d", d.Name, strings.Join(d.Args, " "), d.Line)

				if value, has := manager.DealWith(valueType, item); has {
					m.SetMapIndex(key, value)
				} else {
					vs := reflect.New(valueType)
					if vs.Kind() == reflect.Ptr {
						vs = vs.Elem()
					}
					value, err := assemblyValue(valueType, vs, d, manager)
					utils.Panic(err, "invalid %s %s, line %d", item.Name, strings.Join(item.Args, " "), item.Line)
					m.SetMapIndex(key, value)
				}
			}
		}
	}
	return m
}

func sliceValue(sliceType reflect.Type, item *config.Directive, manager TypeManager) reflect.Value {
	if isBase(sliceType) {
		length := len(item.Args)
		values := reflect.MakeSlice(reflect.SliceOf(sliceType), length, length)
		for i, arg := range item.Args {
			if v, err := baseValue(sliceType, arg, manager); err == nil {
				values.Index(i).Set(v)
			}
		}
		return values
	} else {
		slice := reflect.MakeSlice(reflect.SliceOf(sliceType), 1, 1)
		if value, has := manager.DealWith(sliceType, item); has {
			slice.Index(0).Set(value)
		} else {
			vs := reflect.New(sliceType)
			if vs.Kind() == reflect.Ptr {
				vs = vs.Elem()
			}
			if value, err := assemblyValue(sliceType, vs, item, manager); err == nil {
				slice.Index(0).Set(value)
			}
		}
		return slice
	}
}

//OK
func baseValue(fieldType reflect.Type, value string, manager TypeManager) (reflect.Value, error) {

	if fieldType.Kind() == reflect.Ptr {
		if out, err := baseValue(fieldType.Elem(), value, manager); err == nil {
			v := reflect.New(fieldType.Elem())
			v.Elem().Set(out)
			return v, err
		} else {
			return out, err
		}
	}

	v := reflect.New(fieldType)
	switch fieldType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := strconv.ParseInt(value, 10, 64); err != nil {
			return v, err
		} else {
			v.Elem().SetInt(i)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if i, err := strconv.ParseUint(value, 10, 64); err != nil {
			return v, err
		} else {
			v.Elem().SetUint(i)
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(value, 64); err != nil {
			return v, err
		} else {
			v.Elem().SetFloat(f)
		}
	case reflect.Bool:
		v.Elem().SetBool(value == "" || value == "true")

	case reflect.String:
		v.Elem().SetString(value)
	case reflect.Struct:
		if fieldType.String() != "time.Time" {
			return v, os.ErrInvalid
		}
		if t, err := time.Parse("2006-01-02 15:04:05", value); err != nil {
			return v, err
		} else {
			v.Elem().Set(reflect.ValueOf(t))
		}
	default:
		return v, os.ErrInvalid
	}
	return v.Elem(), nil
}
