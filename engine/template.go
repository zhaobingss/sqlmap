package engine

import (
	"io"
	"text/template"
)

/// 模板构建器
type TemplateBuilder interface {
	New(string, string) (Template, error)
}

/// 模板执行器
type TemplateExecutor interface {
	Execute(io.Writer, interface{}) error
}

/// 模板聚合
type Template interface {
	TemplateBuilder
	TemplateExecutor
}

/// 默认的模板
type DefaultTemplate struct {
	tpl *template.Template
}

func (dt *DefaultTemplate) New(name, content string) (Template, error) {
	tpl := template.New(name)
	tpl, err := tpl.Parse(content)
	if err != nil {
		return nil, err
	}
	dt.tpl = tpl
	return dt, nil
}

func (dt *DefaultTemplate) Execute(wr io.Writer, param interface{}) error {
	return dt.tpl.Execute(wr, param)
}