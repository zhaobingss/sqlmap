package engine

import (
	"io"
	"text/template"
)

/// 模板构建器
type TemplateBuilder interface {
	/// 1、模板的名字 2、模板的内容
	New(string, string) (Template, error)
}

/// 模板执行器
type TemplateExecutor interface {
	/// 1、解析内容的容器 2、传入模板的参数
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
	dtp := &DefaultTemplate{
		tpl:tpl,
	}
	return dtp, nil
}

func (dt *DefaultTemplate) Execute(wr io.Writer, param interface{}) error {
	return dt.tpl.Execute(wr, param)
}
