package engine

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"github.com/zhaobingss/sqlmap/util"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"
)

/// sql文件默认命名空间
var DefaultNamespace = "default_namespace"
/// 存储sql预处理语句
var SqlMap = map[string]string{}
/// 存储sql处理模板
var TplMap = map[string]*Template{}
/// 处理sql多余空格的正则
var reg, _ = regexp.Compile("\\s+")

/// sql引擎
type SqlEngine struct {
	lock           sync.RWMutex
	db             *sql.DB
	init           bool
	tplType        string
	sessionFactory *SessionFactory
}

func New() *SqlEngine {
	engine := &SqlEngine{}
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
	s.sessionFactory = NewSessionFactory(s.db)
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

/// 获取session
func (s *SqlEngine) NewSession() *Session {
	s.checkInit()
	return s.sessionFactory.NewSession()
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

/// 构建sql语句
func (s *SqlEngine) buildSql(key string, data interface{}) (string, error) {
	s.checkInit()
	content := SqlMap[key]
	if content == "" {
		return "", errors.New("未匹配到映射：" + key)
	}

	tpl, err := s.getAndSetTemplate(key, content, s.tplType)
	if err != nil {
		return "", err
	}
	bts := &bytes.Buffer{}
	err = tpl.Execute(bts, data)
	val := bts.String()
	val = strings.TrimSpace(val)
	val = reg.ReplaceAllString(val, " ")
	return val, err
}

/// 获取和设置模板
func (s *SqlEngine) getAndSetTemplate(key, content, typ string) (*Template, error) {
	tpl := TplMap[key]
	var err error
	if tpl == nil {
		tpl, err = NewTemplate(key, content, typ)
		if err != nil {
			return nil, err
		}
		s.lock.Lock()
		TplMap[key] = tpl
		s.lock.Unlock()
	}
	return tpl, nil
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

	m, err := s.parse(bts)
	if err != nil {
		return err
	}
	if m != nil && len(m) > 0 {
		for k, v := range m {
			SqlMap[k] = v
		}
	}

	return nil
}

/// 解析xml文件处理成sql语句
func (s *SqlEngine) parse(xml []byte) (map[string]string, error) {
	ret := map[string]string{}

	doc := etree.NewDocument()
	err := doc.ReadFromBytes(xml)
	if err != nil {
		return ret, err
	}

	sm := doc.SelectElement("sqlmap")
	if sm == nil {
		return nil, errors.New("缺少sqlmap节点")
	}

	namespace := sm.SelectAttrValue("namespace", DefaultNamespace)
	if namespace == "" {
		namespace = DefaultNamespace
	}

	els := sm.SelectElements("sql")
	if els == nil || len(els) < 1 {
		return ret, nil
	}

	for _, e := range els {
		id := e.SelectAttrValue("id", "")
		if id == "" {
			return ret, errors.New(namespace + " 中有sql语句未设置ID")
		}
		fullId := namespace + "_" + id
		if ret[fullId] == fullId {
			return ret, errors.New(namespace + " 中 " + fullId + " 重复")
		}
		val := e.Text()
		val = strings.Replace(val, "\n", " ", -1)
		val = strings.Trim(val, "\n")
		val = strings.TrimSpace(val)
		ret[fullId] = val
	}

	return ret, nil
}

/// 检测引擎是否初始化
func (s *SqlEngine) checkInit() {
	if !s.init {
		panic(errors.New("未初始化引擎"))
	}
}


