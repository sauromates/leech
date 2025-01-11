package utils

import "reflect"

// FlattenTaggedStruct uses reflection to unwrap any embedded structs from
// given value and returns a flat map of all tagged fields.
//
// Any embedded structs will be flattened recursively with same logic.
//
// If `val` is a pointer, it will be dereferenced to a value. If `val` is not
// a struct, an empty map will be returned.
func FlattenTaggedStruct(val any, tag string) map[string]any {
	flat := make(map[string]any)
	v, t := reflect.ValueOf(val), reflect.TypeOf(val)

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return flat
	}

	for i := range v.NumField() {
		field, fieldType := v.Field(i), t.Field(i)
		isEmbedded := field.Kind() == reflect.Struct && fieldType.Anonymous

		if isEmbedded {
			flatEmbedded := FlattenTaggedStruct(field.Interface(), tag)
			for k, vv := range flatEmbedded {
				flat[k] = vv
			}
		} else {
			fieldTag := fieldType.Tag.Get(tag)
			if fieldTag != "" && field.CanInterface() {
				flat[fieldTag] = field.Interface()
			}
		}
	}

	return flat
}
