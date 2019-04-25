package engine

import (
	"database/sql"
	"errors"
	"github.com/beevik/etree"
	"github.com/zhaobingss/sqlmap/util"
	"io/ioutil"
	"strings"
	"sync"
)

/// sql引擎
type SqlEngine struct {
	lock           sync.RWMutex
	db             *sql.DB
	init           bool
	tplType        string
	sessionFactory *sessionFactory
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
	var err error
	s.db, err = sql.Open(driver, dataSrcName)
	if err != nil {
		return err
	}
	s.tplType = typ
	if typ == "" {
		s.tplType = TplDefault
	}
	s.sessionFactory = newSessionFactory(s.db)
	err = s.initSql(sqlDir)

	return err
}

/// 执行非SELECT的sql
func (s *SqlEngine) Execute(key string, data interface{}) (sql.Result, error) {
	s.checkInit()
	return exec(key, s.tplType, data, s.db.Exec)
}

/// 执行SELECT的sql
func (s *SqlEngine) Query(key string, data interface{}) ([]map[string]string, error) {
	s.checkInit()
	return query(key, s.tplType, data, s.db.Prepare)
}

/// 获取session
func (s *SqlEngine) NewSession() *Session {
	s.checkInit()
	return s.sessionFactory.newSession(s.tplType)
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
			vv := SqlMap[k]
			if vv != "" {
				return errors.New("映射KEY重复：" + k)
			} else {
				SqlMap[k] = v
			}
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
		fullId := namespace + "." + id
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
