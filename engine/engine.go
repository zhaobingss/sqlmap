package engine

import (
	"database/sql"
	"errors"
	"github.com/beevik/etree"
	"github.com/zhaobingss/sqlmap/log"
	"github.com/zhaobingss/sqlmap/util"
	"io/ioutil"
	"strings"
	"sync"
)

/// the SqlEngine
type SqlEngine struct {
	lock sync.RWMutex
	db   *sql.DB
	init bool
}

/// create a new engine without init
func New() *SqlEngine {
	engine := &SqlEngine{}
	return engine
}

/// create a new engine with init
/// @param driver: db drive name, eg: mysql,sqlite
/// @param dataSrcName: eg: root:root@(127.0.0.1:3306)/test
/// @param sqlDir: the sql.goxml files dir
func NewEngine(driver, dataSrcName, sqlDir string) (*SqlEngine, error) {
	engine := New()
	err := engine.Init(driver, dataSrcName, sqlDir)
	return engine, err
}

/// get the database/sql.DB
func (s *SqlEngine) GetDB() *sql.DB {
	s.checkInit()
	return s.db
}

/// init the sql engine
/// @param driver: db drive name, eg: mysql,sqlite
/// @param dataSrcName: eg: root:root@(127.0.0.1:3306)/test
/// @param sqlDir: the sql.goxml files dir
func (s *SqlEngine) Init(driver, dataSrcName, sqlDir string) error {
	if s.init {
		return errors.New("the engine is already init")
	}
	defer func() {
		s.init = true
	}()
	var err error
	s.db, err = sql.Open(driver, dataSrcName)
	if err != nil {
		return err
	}
	err = s.initSql(sqlDir)
	tplBuilder = &DefaultTemplate{}
	return err
}

/// execute the sql with a can ignore result
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
func (s *SqlEngine) Execute(key string, param interface{}) (sql.Result, error) {
	s.checkInit()
	return exec(key, param, s.db.Exec)
}

/// execute the sql and set result to []map[string]string
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
func (s *SqlEngine) Query(key string, param interface{}) ([]map[string]string, error) {
	s.checkInit()
	return query(key, param, s.db.Query)
}

/// execute sql and set the result to a slice dest
/// @param the result will be set to dest, and the dest must be like eg: *[]*struct or *[]struct
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
func (s *SqlEngine) Select(dest interface{}, key string, param interface{}) error {
	s.checkInit()
	return selectRows(dest, key, param, s.db.Query)
}

/// execute sql and set the result to a struct dest
/// @param the result will be set to dest, and the dest must be like eg: *struct
/// @param key: sql map key, namespace + sql ID
/// @param param: the param to pass to the sql template
/// @return error: ERR_NOT_GOT_RECORD,ERR_MORE_THAN_ONE_RECORD,...
/// ERR_NOT_GOT_RECORD indicate that not got any recode from the database
/// ERR_MORE_THAN_ONE_RECORD indicate that got more than one record from database
func (s *SqlEngine) SelectOne(dest interface{}, key string, param interface{}) error {
	s.checkInit()
	return selectRow(dest, key, param, s.db.Query)
}

/// get a session use for transaction
func (s *SqlEngine) NewSession() *Session {
	s.checkInit()
	return newSession(s.db)
}

/// register a sql template to replace the default, default use go text/template
/// @param tb: the template builder
func (s *SqlEngine) RegisterTemplate(tb TemplateBuilder) {
	tplBuilder = tb
}

/// register the log func
/// @param err: the error log func
/// @param inf: the info log func
func (s *SqlEngine) RegisterLogFunc(err, inf func(f interface{}, v ...interface{})) {
	log.RegisterLogFunc(err, inf)
}

/// init the *.goxml files to map
/// @param sqlDir: the *.goxml file location
func (s *SqlEngine) initSql(sqlDir string) error {
	files, err := util.GetFiles(sqlDir)
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

/// load the *goxml file content to map
/// @param file: the *.goxml file path
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
			vv := sqlMap[k]
			if vv != nil {
				return errors.New("the *.goxml map key is repeat: " + k)
			} else {
				sqlMap[k] = &SqlTemplate{sql: v}
			}
		}
	}

	return nil
}

/// parse the *.goxml file
func (s *SqlEngine) parse(xml []byte) (map[string]string, error) {
	ret := map[string]string{}

	doc := etree.NewDocument()
	err := doc.ReadFromBytes(xml)
	if err != nil {
		return ret, err
	}

	sm := doc.SelectElement("sqlmap")
	if sm == nil {
		return nil, errors.New("the sqlmap element is not found")
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
			return ret, errors.New(namespace + " has sql not have ID")
		}
		fullId := namespace + "." + id
		if ret[fullId] == fullId {
			return ret, errors.New(namespace + "." + fullId + " repeat")
		}
		val := e.Text()
		val = strings.Replace(val, "\n", " ", -1)
		val = strings.Trim(val, "\n")
		val = strings.TrimSpace(val)
		ret[fullId] = val
	}

	return ret, nil
}

/// check if the engine is init
func (s *SqlEngine) checkInit() {
	if !s.init {
		panic(errors.New("the engine is not initial"))
	}
}
