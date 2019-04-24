package engine

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/zhaobingss/sqlmap/builder"
	"github.com/zhaobingss/sqlmap/parser"
	"github.com/zhaobingss/sqlmap/util"
	"io/ioutil"
)

type SqlEngine struct {
	db         *sql.DB
	sqlBuilder *builder.SqlBuilder
	xmlParser  *parser.XmlParser
	init       bool
	tplType    string
	sqlMap     map[string]string
}

func New() *SqlEngine {
	engine := &SqlEngine{}
	engine.sqlBuilder = builder.New()
	engine.xmlParser = parser.New()
	engine.sqlMap = map[string]string{}
	return engine
}

/// 获取database/sql.DB
func (s *SqlEngine) GetDB() *sql.DB {
	s.checkInit()
	return s.db
}

/// 初始化sql引擎
func (s *SqlEngine) Init(driver, dataSrcName, sqlDir, typ string) error {
	defer func() {
		s.init = true
	}()
	s.tplType = typ
	var err error
	s.db, err = sql.Open(driver, dataSrcName)
	if err != nil {
		return err
	}

	err = s.initSql(sqlDir)

	return err
}

/// 执行非SELECT的sql
func (s *SqlEngine) Execute(key string, data interface{}) (int64, error) {
	s.checkInit()
	sqlStr, err := s.buildSql(key, data)
	if err != nil {
		return -1, err
	}
	fmt.Println(sqlStr)

	result, err := s.db.Exec(sqlStr)
	if err != nil {
		return -1, err
	}
	return result.LastInsertId()

}

/// 执行SELECT的sql
func (s *SqlEngine) Query(key string, data interface{}) ([]map[string]string, error) {
	s.checkInit()
	sqlStr, err := s.buildSql(key, data)
	if err != nil {
		return nil, err
	}

	fmt.Println(sqlStr)

	stmt, err := s.db.Prepare(sqlStr)
	defer s.closeStmt(stmt)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query()
	defer s.closeRows(rows)
	if err != nil {
		return nil, err
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	size := len(cols)
	rs := make([][]interface{}, 0)
	for rows.Next() {
		r := s.makeEmptyRow(size)
		err = rows.Scan(r...)
		if err != nil {
			return nil, err
		} else {
			rs = append(rs, r)
		}
	}

	m := s.convertMap(rs, cols)

	return m, nil
}

/// 关闭预编译语句
func (s *SqlEngine) closeStmt(stmt *sql.Stmt) {
	if stmt != nil {
		err := stmt.Close()
		if err != nil {
			fmt.Println(err)
		}
	}
}

/// 关闭rows释放连接
func (s *SqlEngine) closeRows(rows *sql.Rows) {
	if rows != nil {
		err := rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}
}

/// 根据sql的id（namespae+sqlId）构建sql
/// key=namespace+"_"+sqlid
/// data 传入本条sql的数据，使用text/template来构建sql
func (s *SqlEngine) buildSql(key string, data interface{}) (string, error) {
	s.checkInit()

	val := s.sqlMap[key]
	if val == "" {
		return "", errors.New("未匹配到映射：" + key)
	}
	return s.sqlBuilder.BuildSql(key, val, s.tplType, data)
}

/// 将返回结果转换为map[string][string]
func (s *SqlEngine) convertMap(rows [][]interface{}, cols []string) []map[string]string {
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

/// 构造空行数据
func (s *SqlEngine) makeEmptyRow(colSize int) []interface{} {
	row := make([]interface{}, colSize)
	for i := 0; i < colSize; i++ {
		row[i] = &[]byte{}
	}
	return row
}

/// 初始化sql语句到内存
func (s *SqlEngine) initSql(sqlDir string) error {
	files, err := util.GetAllFiles(sqlDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		err = s.initSqlMap(f)
		if err != nil {
			return err
		}
	}

	return nil
}

/// 初始化sql映射
func (s *SqlEngine) initSqlMap(file string) error {
	bts, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	m, err := s.xmlParser.Parse(bts)
	if err != nil {
		return err
	}
	if m != nil && len(m) > 0 {
		for k, v := range m {
			s.sqlMap[k] = v
		}
	}

	return nil
}

/// 检测引擎是否初始化
func (s *SqlEngine) checkInit() {
	if !s.init {
		panic(errors.New("未初始化引擎"))
	}
}
