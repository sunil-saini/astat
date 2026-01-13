package model

import "reflect"

func ToAnySlice(v any) []any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil
	}
	l := rv.Len()
	res := make([]any, l)
	for i := range l {
		res[i] = rv.Index(i).Interface()
	}
	return res
}
