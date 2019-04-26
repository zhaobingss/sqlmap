package engine

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

/// sql文件默认命名空间
var DefaultNamespace = "default_namespace"
/// 存储sql预处理语句
var SqlMap = map[string]*SqlTemplate{}
/// 处理sql多余空格的正则
var reg, _ = regexp.Compile("\\s+")
/// 模板构建器
var tplBuilder TemplateBuilder

/// sql内容和处理内容的模板
type SqlTemplate struct {
	sql string   // sql内容
	tpl Template // 处理模板
}

/// 将sql.Rows 转换成 []map[string]string
func convertRows2SliceMapString(rows *sql.Rows) ([]map[string]string, error) {
	rs, cols, err := convertRows2SliceInterface(rows)
	if err != nil {
		return nil, err
	}
	ret := convertMapString(rs, cols)
	return ret, nil
}

/// 将sql.Rows 转换成 []map[string][]byte
func convertRows2SliceMapBytes(rows *sql.Rows) ([]map[string][]byte, error) {
	rs, cols, err := convertRows2SliceInterface(rows)
	if err != nil {
		return nil, err
	}
	ret := convertMapBytes(rs, cols)
	return ret, nil
}

/// 将sql.Rows 转换成 [][]interface{}
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

/// 构造空行用来接收数据库数据
func makeEmptyRow(colSize int) []interface{} {
	row := make([]interface{}, colSize)
	for i := 0; i < colSize; i++ {
		row[i] = &[]byte{}
	}
	return row
}

/// 将[][]interface{}转换为[]map[string][]byte
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

/// 将[][]interface{}转换为[]map[string]string
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

/// 构建sql语句
func buildSql(key string, param interface{}) (string, error) {
	mapper := SqlMap[key]
	if mapper == nil {
		return "", errors.New("未匹配到映射：" + key)
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

/// 获取和设置模板
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

/// 查询sql
func query(key string, param interface{}, f func(string,...interface{}) (*sql.Rows, error)) ([]map[string]string, error) {
	sqlStr, err := buildSql(key, param)
	if err != nil {
		return nil, err
	}

	fmt.Println(sqlStr)

	rows, err := f(sqlStr)
	if err != nil {
		return nil, err
	}

	m, err := convertRows2SliceMapString(rows)
	if err != nil {
		return nil, err
	}

	return m, nil
}

/// 非查询sql
func exec(key string, param interface{}, f func(string, ...interface{}) (sql.Result, error)) (sql.Result, error) {

	sqlStr, err := buildSql(key, param)
	if err != nil {
		return nil, err
	}
	fmt.Println(sqlStr)

	result, err := f(sqlStr)
	if err != nil {
		return nil, err
	}
	return result, nil

}
