package flags

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	flag "github.com/spf13/pflag"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type FlagTemplateFun func(flagName, value string) string

func Flags(flags *flag.FlagSet, main interface{}, prefix string, tempFn ...FlagTemplateFun) {
	typ := reflect.TypeOf(main)
	utils.Assert(typ.Kind() == reflect.Ptr, "value must be pointer to struct, but is %s", typ.Kind())

	mainVal := reflect.ValueOf(main).Elem()
	mainTyp := mainVal.Type()
	utils.Assert(mainTyp.Kind() == reflect.Struct, "value must be pointer to struct, but is pointer to %s", typ.Kind())

	utils.Panic(setFlags(flags, main, prefix, tempFn...), "set flags %v", typ)
}

func tag(field reflect.StructField, names ...string) string {
	for _, name := range names {
		if value, ok := field.Tag.Lookup(name); ok {
			return value
		}
	}
	return ""
}

func templateValue(flagName, value string, tempFn ...FlagTemplateFun) string {
	if value == "" || !strings.Contains(value, "{{") {
		return value
	}
	if len(tempFn) > 0 {
		for _, fun := range tempFn {
			if newV := fun(flagName, value); newV != value {
				return newV
			}
		}
	}
	return value
}

func setFlags(flags *flag.FlagSet, main interface{}, prefix string, tempFn ...FlagTemplateFun) error {
	mainVal := reflect.ValueOf(main).Elem()
	mainTyp := mainVal.Type()

	for i := 0; i < mainTyp.NumField(); i++ {
		fieldType := mainTyp.Field(i)
		fieldVal := mainVal.Field(i)

		if fieldType.PkgPath != "" {
			continue
		}

		flagName, has := fieldType.Tag.Lookup("flag")
		if flagName == "-" {
			continue
		} else if !has && flagName == "" {
			flagName = utils.Name(fieldType.Name, "")
		}

		if prefix != "" {
			if flagName == "" {
				flagName = prefix
			} else {
				flagName = prefix + "." + flagName
			}
		}

		shorthand := tag(fieldType, "short")
		help := templateValue(flagName, tag(fieldType, "help", "h"), tempFn...)
		defValue := templateValue(flagName, tag(fieldType, "def"), tempFn...)

		switch fieldVal.Interface().(type) {
		case time.Duration:
			p := fieldVal.Addr().Interface().(*time.Duration)
			value := time.Duration(fieldVal.Int())
			if defValue != "" {
				if d, err := time.ParseDuration(defValue); err == nil {
					value = d
				}
			}
			flags.DurationVarP(p, flagName, shorthand, value, help)
			continue
		case net.IP:
			p := fieldVal.Addr().Interface().(*net.IP)
			value := net.IP(fieldVal.Bytes())
			if defValue != "" {
				value = net.ParseIP(defValue)
			}
			flags.IPVarP(p, flagName, shorthand, value, help)
			continue
		case []net.IP:
			p := fieldVal.Addr().Interface().(*[]net.IP)
			flags.IPSliceVarP(p, flagName, shorthand, *p, help)
			continue
		case []string:
			p := fieldVal.Addr().Interface().(*[]string)
			value := *p
			if defValue != "" {
				value = strings.Split(defValue, ",")
			}
			if value == nil {
				value = make([]string, 0)
			}
			flags.StringSliceVarP(p, flagName, shorthand, value, help)
			continue
		}

		// now check basic kinds
		switch fieldType.Type.Kind() {
		case reflect.String:
			p := fieldVal.Addr().Interface().(*string)
			value := *p
			if defValue != "" {
				value = defValue
			}
			flags.StringVarP(p, flagName, shorthand, value, help)
		case reflect.Bool:
			p := fieldVal.Addr().Interface().(*bool)
			value := fieldVal.Bool()
			if defValue != "" {
				value, _ = strconv.ParseBool(defValue)
			}
			flags.BoolVarP(p, flagName, shorthand, value, help)
		case reflect.Int:
			p := fieldVal.Addr().Interface().(*int)
			val := int(fieldVal.Int())
			if defValue != "" {
				val, _ = strconv.Atoi(defValue)
			}
			flags.IntVarP(p, flagName, shorthand, val, help)
		case reflect.Int64:
			p := fieldVal.Addr().Interface().(*int64)
			val := *p
			if defValue != "" {
				val, _ = strconv.ParseInt(defValue, 10, 64)
			}
			flags.Int64VarP(p, flagName, shorthand, val, help)
		case reflect.Float64:
			p := fieldVal.Addr().Interface().(*float64)
			value := *p
			if defValue != "" {
				value, _ = strconv.ParseFloat(defValue, 64)
			}
			flags.Float64VarP(p, flagName, shorthand, value, help)
		case reflect.Uint:
			p := fieldVal.Addr().Interface().(*uint)
			val := *p
			if defValue != "" {
				v, _ := strconv.ParseUint(defValue, 10, 64)
				val = uint(v)
			}
			flags.UintVarP(p, flagName, shorthand, val, help)
		case reflect.Uint64:
			p := fieldVal.Addr().Interface().(*uint64)
			val := *p
			if defValue != "" {
				val, _ = strconv.ParseUint(defValue, 10, 64)
			}
			flags.Uint64VarP(p, flagName, shorthand, val, help)
		case reflect.Slice:
			switch fieldType.Type.Elem().Kind() {
			case reflect.String:
				p := fieldVal.Addr().Interface().(*[]string)
				val := *p
				if defValue != "" {
					val = strings.Split(defValue, ",")
					for i, s := range val {
						val[i] = strings.TrimSpace(s)
					}
				}
				if val == nil {
					val = make([]string, 0)
				}
				flags.StringSliceVarP(p, flagName, shorthand, val, help)
			default:
				return fmt.Errorf("encountered unsupported slice type/kind: %#v at %s", fieldVal, prefix)
			}
		case reflect.Struct:
			newPrefix := flagName
			if strings.HasSuffix(flagName, "!embed") {
				newPrefix = prefix
			}
			err := setFlags(flags, fieldVal.Addr().Interface(), newPrefix, tempFn...)
			if err != nil {
				return err
			}
		case reflect.Map:
			p := fieldVal.Addr().Interface().(*map[string]string)
			val := *p
			if defValue != "" {
				val = make(map[string]string)
				for _, pair := range strings.Split(defValue, ",") {
					kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
					if len(kv) == 1 {
						val[strings.TrimSpace(kv[0])] = ""
					} else {
						val[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
					}
				}
			}
			if val == nil {
				val = make(map[string]string)
			}
			flags.StringToStringVarP(p, flagName, shorthand, val, help)
		default:
			return fmt.Errorf("encountered unsupported field type/kind: %#v at %s", fieldVal, prefix)
		}
	}
	return nil
}
