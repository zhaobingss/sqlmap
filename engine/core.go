package engine

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/zhaobingss/sqlmap/log"
	"reflect"
	"regexp"
	"strings"
)

/// the sql default namespace
var DefaultNamespace = "default_namespace"

/// cache sql template
var sqlMap = map[string]*SqlTemplate{}

/// the regex to be use replace space char in sql
var reg, _ = regexp.Compile("\\s+")

/// the template builder instance
var tplBuilder TemplateBuilder

/// a temp var to receive the value that not define in struct
var tmpVar = &[]byte{}

/// the sql and sql template
type SqlTemplate struct {
	sql string   // sql content
	tpl Template // template for generate the execute sql
}

/// convert sql.Rows to []map[string]string
/// @param rows: *sql.Rows
/// @return []map[string]string
/// @return error
func convertRows2SliceMapString(rows *sql.Rows) ([]map[string]string, error) {
	rs, cols, err := convertRows2SliceInterface(rows)
	if err != nil {
		return nil, err
	}
	ret := convertMapString(rs, cols)
	return ret, nil
}

/// convert sql.Rows to []map[string][]byte
/// @param rows: *sql.Rows
/// @return []map[string][]byte
/// @return error
func convertRows2SliceMapBytes(rows *sql.Rows) ([]map[string][]byte, error) {
	rs, cols, err := convertRows2SliceInterface(rows)
	if err != nil {
		return nil, err
	}
	ret := convertMapBytes(rs, cols)
	return ret, nil
}

/// convert sql.Rows to [][]interface{}
/// @param rows: *sql.Rows
/// @return [][]interface{}
/// @return error
func convertRows2SliceInterface(rows *sql.Rows) ([][]interface{}, []string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	size := len(cols)
	rs := make([][]interface{}, 0)
	for rows.Next() {
		r := makeEmptyRow(size)
		err = rows.Scan(r...)
		if err != nil {
			return nil, nil, err
		} else {
			rs = append(rs, r)
		}
	}
	return rs, cols, nil
}

/// make a empty row to receive data from sql.Rows
/// @param colSize: the sql.Rows columns size
func makeEmptyRow(colSize int) []interface{} {
	row := make([]interface{}, colSize)
	for i := 0; i < colSize; i++ {
		row[i] = &[]byte{}
	}
	return row
}

/// convert [][]interface{} to []map[string][]byte
func convertMapBytes(rows [][]interface{}, cols []string) []map[string][]byte {
	ret := make([]map[string][]byte, 0)
	for _, row := range rows {
		m := map[string][]byte{}
		for i, col := range cols {
			if row[i] == nil {
				m[col] = nil
			} else {
				btr := row[i].(*[]byte)
				m[col] = *btr
			}
		}
		ret = append(ret, m)
	}
	return ret
}

/// convert [][]interface{} to []map[string]string
func convertMapString(rows [][]interface{}, cols []string) []map[string]string {
	ret := make([]map[string]string, 0)
	for _, row := range rows {
		m := map[string]string{}
		for i, col := range cols {
			if row[i] == nil {
				m[col] = ""
			} else {
				btr := row[i].(*[]byte)
				m[col] = string(*btr)
			}
		}
		ret = append(ret, m)
	}
	return ret
}

/// build sql for execute
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
func buildSql(key string, param interface{}) (string, error) {
	if key == "" {
		return "", errors.New("the map key must be not empty")
	}
	mapper := sqlMap[key]
	if mapper == nil {
		return "", errors.New("can't match the map key: " + key)
	}

	tpl, err := getAndSetTemplate(key, mapper)
	if err != nil {
		return "", err
	}
	bts := &bytes.Buffer{}
	err = tpl.Execute(bts, param)
	val := bts.String()
	val = strings.TrimSpace(val)
	val = reg.ReplaceAllString(val, " ")
	return val, err
}

/// get or set sql template
/// @param key: sql map key, namespace + sql ID
/// @param mapper: SqlTemplate that store the sql map to Template
func getAndSetTemplate(key string, mapper *SqlTemplate) (Template, error) {
	tpl := mapper.tpl
	var err error
	if tpl == nil {
		tpl, err = tplBuilder.New(key, mapper.sql)
		if err != nil {
			return nil, err
		} else {
			mapper.tpl = tpl
		}
	}
	return tpl, nil
}

/// query and fill the result to []map[string]string
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
/// @param f: the execute func like eg: db.Query/db.Eexcute
/// @return []map[string]string
/// @return error
func query(key string, param interface{}, f func(string, ...interface{}) (*sql.Rows, error)) ([]map[string]string, error) {
	rows, err := queryRows(key, param, f)
	if err != nil {
		return nil, err
	}
	m, err := convertRows2SliceMapString(rows)
	if err != nil {
		return nil, err
	}
	return m, nil
}

/// execute sql
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
/// @param f: the execute func like eg: db.Query/db.Eexcute
/// @return sql.Result
/// @return error
func exec(key string, param interface{}, f func(string, ...interface{}) (sql.Result, error)) (sql.Result, error) {
	sqlStr, err := buildSql(key, param)
	if err != nil {
		return nil, err
	}

	result, err := f(sqlStr)
	if err != nil {
		return nil, err
	}
	return result, nil
}

/// query and fill the result to *[]struct or *[]*struct
/// @param dest: the slice struct that the rows will be set eg: *[]struct or *[]*struct
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
/// @param f: the execute func like eg: db.Query/db.Eexcute
/// @return error
func selectRows(dest interface{}, key string, param interface{}, f func(string, ...interface{}) (*sql.Rows, error)) error {
	rows, err := queryRows(key, param, f)
	if err != nil {
		return err
	}
	err = scanRows(dest, rows)
	return err
}

/// fill the result to *struct
/// @param dest: the struct that the rows will be set eg: *struct
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
/// @param f: the execute func like eg: db.Query/db.Eexcute
/// @return error
func selectRow(dest interface{}, key string, param interface{}, f func(string, ...interface{}) (*sql.Rows, error)) error {
	rows, err := queryRows(key, param, f)
	if err != nil {
		return err
	}
	return scanRow(dest, rows)
}

/// query rows
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
/// @param f: the execute func like eg: db.Query/db.Eexcute
/// @return *sql.Rows
/// @return error
func queryRows(key string, param interface{}, f func(string, ...interface{}) (*sql.Rows, error)) (*sql.Rows, error) {
	sqlStr, err := buildSql(key, param)
	if err != nil {
		return nil, err
	}

	if log.Info != nil {
		log.Info(sqlStr)
	}

	rows, err := f(sqlStr)
	return rows, err
}

/// set the result set to slice struct
/// @param dest: the slice struct that the rows will be set eg: *[]struct or *[]*struct
/// @param *sql.Rows
func scanRows(dest interface{}, rows *sql.Rows) error {
	val := reflect.ValueOf(dest)
	err := checkScanRowsType(val.Type())
	if err != nil {
		return err
	}
	direct := reflect.Indirect(val)

	slice, err := detectBaseType(val.Type(), reflect.Slice)
	if err != nil {
		return err
	}

	isPtr := slice.Elem().Kind() == reflect.Ptr
	base := deRefType(slice.Elem())
	columns, err := rows.Columns()
	for rows.Next() {
		vp := reflect.New(base)
		v := reflect.Indirect(vp)
		fields, err := makeReflectRow(v, columns)
		if err != nil {
			return err
		}
		err = rows.Scan(fields...)
		if err != nil {
			return err
		}
		if isPtr {
			direct.Set(reflect.Append(direct, vp))
		} else {
			direct.Set(reflect.Append(direct, v))
		}
	}
	return nil
}

/// set the result set to struct or struct pointer
/// @param dest: the struct that the rows will be set eg: *struct
/// @param *sql.Rows
func scanRow(dest interface{}, rows *sql.Rows) error {
	val := reflect.ValueOf(dest)
	err := checkScanRowType(val.Type())
	if err != nil {
		return err
	}

	direct := reflect.Indirect(val)
	structType, err := detectBaseType(val.Type(), reflect.Struct)
	if err != nil {
		return err
	}

	base := deRefType(structType)
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return ERR_NOT_GOT_RECORD
	}

	vp := reflect.New(base)
	v := reflect.Indirect(vp)
	fields, err := makeReflectRow(v, columns)
	if err != nil {
		return err
	}

	err = rows.Scan(fields...)
	if err != nil {
		return err
	}
	if rows.Next() {
		return ERR_MORE_THAN_ONE_RECORD
	}

	direct.Set(v)

	return nil
}

/// make a columns slice to receive the rows scan result
func makeReflectRow(val reflect.Value, columns []string) ([]interface{}, error) {
	fields := make([]interface{}, len(columns))
	for i, column := range columns {
		f, err := chooseReflectField(val, column)
		if err != nil {
			return nil, err
		}
		fields[i] = f.Addr().Interface()
	}
	return fields, nil
}

/// set the struct field to slice column
func chooseReflectField(val reflect.Value, column string) (reflect.Value, error) {
	val = reflect.Indirect(val)
	fieldLen := val.NumField()
	for i := 0; i < fieldLen; i++ {
		fv := val.Field(i)
		tag := val.Type().Field(i).Tag.Get(`db`)
		if tag == column {
			return fv, nil
		}
	}
	return reflect.Indirect(reflect.ValueOf(tmpVar)), nil
}

/// check the dest type
func checkScanRowsType(typ reflect.Type) error {
	if typ.Kind() != reflect.Ptr {
		return errors.New(`must pass a pointer, not a value`)
	}
	if typ.Elem().Kind() != reflect.Slice {
		return errors.New(`the obj must a pointer to slice`)
	}
	if typ.Elem().Elem().Kind() == reflect.Ptr {
		if typ.Elem().Elem().Elem().Kind() != reflect.Struct {
			return errors.New(`the slice item must a struct or struct pointer`)
		} else {
			return nil
		}
	} else {
		if typ.Elem().Elem().Kind() != reflect.Struct {
			return errors.New(`the slice item must a struct or struct pointer`)
		} else {
			return nil
		}
	}
}

/// check the dest type
func checkScanRowType(typ reflect.Type) error {
	if typ.Kind() != reflect.Ptr {
		return errors.New(`must pass a pointer, not a value`)
	} else {
		if typ.Elem().Kind() != reflect.Struct {
			return errors.New(`the param must a struct or struct pointer`)
		} else {
			return nil
		}
	}
}

/// detect the dest finally type not a pointer
func detectBaseType(t reflect.Type, expected reflect.Kind) (reflect.Type, error) {
	t = deRefType(t)
	if t.Kind() != expected {
		return nil, fmt.Errorf("expected %s but got %s", expected, t.Kind())
	}
	return t, nil
}

/// get the dest reflect type not pointer
func deRefType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
