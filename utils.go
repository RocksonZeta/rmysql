package rmysql

import (
	"reflect"
	"strings"
)

func PtrValue(v reflect.Value) reflect.Value {
	if reflect.Ptr == v.Kind() {
		return PtrValue(v.Elem())
	}
	return v
}

func Args(args interface{}) []interface{} {
	argsV := reflect.ValueOf(args)
	argsLen := argsV.Len()
	r := make([]interface{}, argsLen)
	for i := 0; i < argsLen; i++ {
		r[i] = argsV.Index(i)
	}
	return r
}
func ArgsString(fields []string) []interface{} {
	args := make([]interface{}, len(fields))
	for i := 0; i < len(fields); i++ {
		args[i] = fields[i]
	}
	return args
}
func ArgsStringAhead(key string, fields []string) []interface{} {
	args := make([]interface{}, len(fields)+1)
	args[0] = key
	for i := 0; i < len(fields); i++ {
		args[i+1] = fields[i]
	}
	return args
}

func JoinFields(prefix string, fields ...string) string {
	r := ""
	l := len(fields)
	j := strings.Index(prefix, ".")
	p := prefix
	q := ""
	if j != len(prefix)-1 {
		p = prefix[0 : j+1]
		q = prefix[j+1:]
	}
	for i, f := range fields {
		r += p + "`" + strings.Trim(f, q) + "`"
		if q != "" {
			r += " as " + f
		}
		if i != l-1 {
			r += ","
		}
	}
	return r
}

func GetField(v interface{}, field string) interface{} {
	vV := PtrValue(reflect.ValueOf(v))
	return vV.FieldByName(field).Interface()
}

func Unique(ids []int) []int {
	m := make(map[int]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	r := make([]int, len(m))
	i := 0
	for id := range m {
		r[i] = id
		i++
	}
	return r
}
func MapIds(ids []int, fn func(id int) string) []string {
	r := make([]string, len(ids))
	for i, id := range ids {
		r[i] = fn(id)
	}
	return r
}

func Col(arr interface{}, result interface{}, field string) {
	arrV := PtrValue(reflect.ValueOf(arr))
	arrLen := arrV.Len()
	colType := reflect.TypeOf(result).Elem()
	col := reflect.MakeSlice(colType, arrLen, arrLen)
	for i := 0; i < arrLen; i++ {
		ele := arrV.Index(i)
		col.Index(i).Set(ele.FieldByName(field))
	}
	reflect.ValueOf(result).Elem().Set(col)
}
func ColInt2Str(arr interface{}, field string, fn func(id int) string) []string {
	ints := ColInt(arr, field)
	r := make([]string, len(ints))
	for i, v := range ints {
		r[i] = fn(v)
	}
	return r
}

//ColInt 获取结构体数组的整数列。arr:T[]
func ColInt(arr interface{}, field string) []int {
	var r []int
	Col(arr, &r, field)
	return r
}
