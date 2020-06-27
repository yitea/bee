package beegopro

var BeeDefaultControllerTmpl = "bee_default_controller.go.tmpl"
var BeeDefaultModelTmpl = "bee_default_model.go.tmpl"

type ModelSchema struct {
	Name       string // 字段名
	CamelName  string // 驼峰字段名
	Type       string // MySQL中原始数据类型
	ColumnKey  string // PRI说明是主键
	Comment    string // Mysql中原始注释
	GoType     string // Go结构体字段类型
	GoJsonTag  string // GO结构体中json标签
	OrmTag     string // orm tag
	IsListShow bool   // 是否显示
	IsOrm      bool   // 是否显示
}

type BaseSchema struct {
	Name      string // table name
	CamelName string // camel table name
}
