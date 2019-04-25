package engine

import (
	"errors"
	"io"
	"text/template"
)

const (
	DEFAULT = "default"
)

/// 渲染动态sql的模板
type Template struct {
	typ string             // 模板类型
	tpl *template.Template // go默认的模板引擎
}

/// 创建模板
func NewTemplate(name, content, typ string) (*Template, error) {
	if typ == DEFAULT {
		return newDefault(name, content, typ)
	} else {
		return nil, errors.New("不支持的模板类型：[" + typ + "]")
	}
}

/// go使用go的模板引擎
func newDefault(name, content, typ string) (*Template, error) {
	tpl := template.New(name)
	tpl, err := tpl.Parse(content)
	if err != nil {
		return nil, err
	}

	t := &Template{}
	t.typ = typ
	t.tpl = tpl
	return t, nil
}

/// 执行模板渲染
func (t *Template) Execute(wr io.Writer, data interface{}) error {
	if t.typ == DEFAULT {
		return t.tpl.Execute(wr, data)
	} else {
		return errors.New("不支持的模板类型")
	}
}
