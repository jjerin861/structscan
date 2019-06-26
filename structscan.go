package structscan

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type dbStruct struct {
	elem   reflect.Value
	fields map[string][]int
}

func NewDBStruct(dest interface{}) (*dbStruct, error) {
	elem := reflect.ValueOf(dest).Elem()
	fields, err := getFields(elem)
	return &dbStruct{
		elem:   elem,
		fields: fields,
	}, err
}
func (dbs *dbStruct) StructScan(rows *sql.Rows) error {
	var values []interface{}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	for _, name := range cols {
		idx, ok := dbs.fields[strings.ToLower(name)]
		var v interface{}
		if !ok {
			v = &sql.RawBytes{}
		} else {
			v = dbs.elem.FieldByIndex(idx).Addr().Interface()
		}
		values = append(values, v)
	}
	return rows.Scan(values...)
}
func (dbs *dbStruct) Alias(aliases map[string]string) error {
	for k, v := range aliases {
		index, ok := dbs.fields[k]
		if !ok {
			return fmt.Errorf(
				"unable to find db tag: %s",
				k,
			)
		}
		if _, ok := dbs.fields[v]; ok {
			return fmt.Errorf(
				"db tag: %s already allocated",
				v,
			)
		}
		delete(dbs.fields, k)
		dbs.fields[v] = index
	}
	return nil
}
func getFields(elem reflect.Value) (map[string][]int, error) {
	fields := make(map[string][]int)
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Type().Field(i)
		varTag, ok := f.Tag.Lookup("db")
		if !ok {
			return nil, errors.New("struct format is not valid")
		}
		if varTag == "-" {
			return nil, errors.New("should define db tag for every field")
		}
		if f.PkgPath != "" {
			continue
		}
		if _, ok := fields[varTag]; ok {
			return nil, errors.New("morethan one field with same db tag")
		}
		fields[varTag] = []int{i}
	}
	return fields, nil
}
