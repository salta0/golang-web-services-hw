package main

import (
	"errors"
	"reflect"
)

var types []reflect.Type

func i2s(data any, out any) error {
	if reflect.ValueOf(out).Kind() != reflect.Ptr {
		return errors.New("Out not a pointer")
	}

	el := reflect.ValueOf(out).Elem()

	err := fillStruct(data, el)
	if err != nil {
		return err
	}

	return nil
}

func fillStruct(data any, out reflect.Value) error {
	switch out.Kind() {
	case reflect.Struct:
		if reflect.ValueOf(data).Kind() != reflect.Map {
			return errors.New("Invalid data")
		}

		outType := out.Type()
		currentData := data.(map[string]any)

		for i := 0; i < out.NumField(); i++ {
			field := out.Field(i)
			fieldName := outType.Field(i).Name
			switch field.Kind() {
			case reflect.Int:
				if reflect.TypeOf(currentData[fieldName]).String() != "float64" {
					return errors.New("not float64")
				}

				v := currentData[outType.Field(i).Name].(float64)
				field.SetInt(int64(v))
			case reflect.String:
				if reflect.TypeOf(currentData[fieldName]).String() != "string" {
					return errors.New("not string")
				}

				v := currentData[fieldName].(string)
				field.SetString(v)
			case reflect.Bool:
				if reflect.TypeOf(currentData[fieldName]).String() != "bool" {
					return errors.New("not bool")
				}

				v := currentData[fieldName].(bool)
				field.SetBool(v)
			default:
				err := fillStruct(currentData[fieldName], field)
				if err != nil {
					return err
				}
			}
		}
	case reflect.Slice:
		if reflect.ValueOf(data).Kind() != reflect.Slice {
			return errors.New("Invalid data")
		}

		currentData := data.([]any)
		collection := reflect.MakeSlice(out.Type(), len(currentData), len(currentData))

		for i, d := range currentData {
			itemPtr := reflect.New(out.Type().Elem())
			item := reflect.Indirect(itemPtr)
			err := fillStruct(d, item)
			if err != nil {
				return err
			}

			collection.Index(i).Set(item)
		}
		out.Set(collection)
	}

	return nil
}
