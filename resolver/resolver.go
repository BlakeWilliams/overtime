package resolver

import (
	"fmt"
	"reflect"
)

func Resolve(target any, resolver any) error {
	t := reflect.ValueOf(target)

	size := 1
	if t.Kind() == reflect.Slice {
		size = t.Len()

		if size == 0 {
			return nil
		}
	} else if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fields, err := fieldsToResolve(t)
	if err != nil {
		return fmt.Errorf("Failed to get fields to resolve: %w", err)
	}

	targets := make(map[int64]reflect.Value, size)
	ids := make([]int64, size)

	if t.Kind() == reflect.Slice {
		isPtr := t.Index(0).Kind() == reflect.Ptr

		for i := 0; i < t.Len(); i++ {
			target := t.Index(i)
			if isPtr {
				target = target.Elem()
			}
			id := target.FieldByName("ID").Int()
			targets[id] = target
			ids[i] = id
		}
	} else {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		id := t.FieldByName("ID").Int()
		targets[id] = t
		ids[0] = id
	}

	rt := reflect.ValueOf(resolver)
	for fieldName, resolverMethod := range fields {
		resolverMethod := rt.MethodByName(resolverMethod)
		result := resolverMethod.Call([]reflect.Value{reflect.ValueOf(ids)})
		resultMap := result[0]

		err := result[1].Interface()
		if err != nil {
			return fmt.Errorf("Failed to call resolver method: %w", err.(error))
		}

		// Handle nil maps, move onto next resolver
		if resultMap.IsNil() {
			continue
		}

		for id := range targets {
			field := targets[id].FieldByName(fieldName)

			// Ensure the field is valid and settable
			if !field.IsValid() || !field.CanSet() {
				return fmt.Errorf("field %s is not valid or cannot be set", fieldName)
			}

			// Initialize the field if it is zero
			if field.IsZero() {
				switch field.Kind() {
				case reflect.Ptr:
					field.Set(reflect.New(field.Type().Elem()))
				case reflect.Slice:
					field.Set(reflect.MakeSlice(field.Type(), 0, 0))
				case reflect.Map:
					field.Set(reflect.MakeMap(field.Type()))
				case reflect.Struct:
					field.Set(reflect.New(field.Type()).Elem())
				default:
					return fmt.Errorf("unsupported kind %s for field %s", field.Kind(), fieldName)
				}
			}

			if value := resultMap.MapIndex(reflect.ValueOf(id)); value.IsValid() {
				field.Set(value)
			}
		}
	}

	return nil
}

func fieldsToResolve(t reflect.Value) (map[string]string, error) {
	if t.Kind() == reflect.Slice {
		// If there's no elements, there's nothing to resolve
		if t.Len() == 0 {
			return map[string]string{}, nil
		}

		// Get the first element and use that as the target to
		// determine fields to resolve
		t = t.Index(0)

		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	} else if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	reflectType := t.Type()
	resolverMap := make(map[string]string, 0)

	fmt.Println(t.Kind())
	for i := 0; i < t.NumField(); i++ {
		fieldType := reflectType.Field(i)
		resolverMethod := fieldType.Tag.Get("resolver")
		if resolverMethod == "" {
			continue
		}

		resolverMap[fieldType.Name] = resolverMethod
	}

	return resolverMap, nil
}
