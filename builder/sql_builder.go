package builder

import (
	"bytes"
	t "github.com/zhaobingss/sqlmap/template"
	"regexp"
	"strings"
	"sync"
)

var reg, _ = regexp.Compile("\\s+")

type SqlBuilder struct {
	lock sync.RWMutex
	tplMap map[string]*t.Template
}

/// SQL构建器
func New() *SqlBuilder {
	return &SqlBuilder{
		tplMap: map[string]*t.Template{},
	}
}

/// 构建sql
func (s *SqlBuilder) BuildSql(key, content, typ string, data interface{}) (string, error) {
	tpl, err := s.getAndSetTemplate(key, content, typ)
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
func (s *SqlBuilder) getAndSetTemplate(key, content, typ string) (*t.Template, error) {
	tpl := s.tplMap[key]
	var err error
	if tpl == nil {
		tpl, err = t.New(key, content, typ)
		if err != nil {
			return nil, err
		}
		s.lock.Lock()
		s.tplMap[key] = tpl
		s.lock.Unlock()
	}
	return tpl, nil
}
