package parser

import (
	"errors"
	"github.com/beevik/etree"
	"strings"
)

const DEFAULT_NAMESPACE = "default_namespace"

/// xml解析器
type XmlParser struct {
}

func New() *XmlParser {
	return &XmlParser{}
}

/// 解析xml文件处理成sql语句
func (x *XmlParser) Parse(xml []byte) (map[string]string, error) {
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

	namespace := sm.SelectAttrValue("namespace", DEFAULT_NAMESPACE)
	if namespace == "" {
		namespace = DEFAULT_NAMESPACE
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
