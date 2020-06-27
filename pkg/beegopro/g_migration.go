// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package beegopro

import (
	"errors"
	"fmt"
	"github.com/beego/bee/cmd/commands/migrate"
	"strings"
	"time"

	"github.com/beego/bee/logger"
	"github.com/beego/bee/utils"
)

const (
	MDateFormat = "20060102_150405"
)

type DBDriver interface {
	GenerateCreateUp(tableName string, schemas []Schema) string
	GenerateCreateDown(tableName string) string
}

type mysqlDriver struct{}

func (m mysqlDriver) GenerateCreateUp(tableName string, schemas []Schema) string {
	upsql := `m.SQL("CREATE TABLE ` + tableName + "(" + m.generateSQLFromSchemas(schemas) + `)");`
	return upsql
}

func (m mysqlDriver) GenerateCreateDown(tableName string) string {
	downsql := `m.SQL("DROP TABLE ` + "`" + tableName + "`" + `")`
	return downsql
}

func (m mysqlDriver) generateSQLFromSchemas(schemas []Schema) string {
	sql, tags := "", ""
	for i, v := range schemas {
		if v.Orm == "-" {
			continue
		}
		typ, tag := m.getSQLType(v.Type, v.Orm)
		if typ == "" {
			beeLogger.Log.Error("Fields format is wrong. Should be: key:type,key:type " + v.Type)
			return ""
		}
		if i == 0 && v.Type != "autopk" && strings.ToLower(v.Name) != "id" {
			sql += "`id` int(11) NOT NULL AUTO_INCREMENT,"
			tags = tags + "PRIMARY KEY (`id`),"
		}

		if v.Type == "autopk" {
			tags = tags + "PRIMARY KEY (`" + strings.ToLower(v.Name) + "`),"
		}
		sql += "`" + utils.SnakeString(v.Name) + "` " + typ + ","
		if tag != "" {
			tags = tags + fmt.Sprintf(tag, "`"+utils.SnakeString(v.Name)+"`") + ","
		}
	}
	sql = strings.TrimRight(sql+tags, ",")
	return sql
}

func (m mysqlDriver) getSQLType(ktype string, orm string) (tp, tag string) {
	kv := strings.SplitN(ktype, ":", 2)
	switch kv[0] {
	case "string":
		if len(kv) == 2 {
			return "varchar(" + kv[1] + ") NOT NULL", ""
		}
		return "varchar(128) NOT NULL", ""
	case "text":
		return "longtext  NOT NULL", ""
	case "auto":
		return "int(11) NOT NULL AUTO_INCREMENT", ""
	case "autopk":
		return "int(11) NOT NULL AUTO_INCREMENT", ""
	case "pk":
		return "int(11) NOT NULL", "PRIMARY KEY (%s)"
	case "datetime":
		return "datetime NOT NULL", ""
	case "int", "int8", "int16", "int32", "int64":
		fallthrough
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "int(11) DEFAULT NULL", ""
	case "bool":
		return "tinyint(1) NOT NULL", ""
	case "float32", "float64":
		return "float NOT NULL", ""
	case "float":
		return "float NOT NULL", ""
	}
	return "", ""
}

type postgresqlDriver struct{}

func (m postgresqlDriver) GenerateCreateUp(tableName string, schemas []Schema) string {
	upsql := `m.SQL("CREATE TABLE ` + tableName + "(" + m.generateSQLFromSchemas(schemas) + `)");`
	return upsql
}

func (m postgresqlDriver) GenerateCreateDown(tableName string) string {
	downsql := `m.SQL("DROP TABLE ` + tableName + `")`
	return downsql
}

func (m postgresqlDriver) generateSQLFromSchemas(schemas []Schema) string {
	sql, tags := "", ""
	for i, v := range schemas {
		typ, tag := m.getSQLType(v.Type)
		if typ == "" {
			beeLogger.Log.Error("Fields format is wrong. Should be: key:type,key:type " + v.Type)
			return ""
		}
		if i == 0 && strings.ToLower(v.Name) != "id" {
			sql += "id serial primary key,"
		}
		sql += utils.SnakeString(v.Name) + " " + typ + ","
		if tag != "" {
			tags = tags + fmt.Sprintf(tag, utils.SnakeString(v.Name)) + ","
		}
	}
	if tags != "" {
		sql = strings.TrimRight(sql+" "+tags, ",")
	} else {
		sql = strings.TrimRight(sql, ",")
	}
	return sql
}

func (m postgresqlDriver) getSQLType(ktype string) (tp, tag string) {
	kv := strings.SplitN(ktype, ":", 2)
	switch kv[0] {
	case "string":
		if len(kv) == 2 {
			return "char(" + kv[1] + ") NOT NULL", ""
		}
		return "TEXT NOT NULL", ""
	case "text":
		return "TEXT NOT NULL", ""
	case "auto", "pk":
		return "serial primary key", ""
	case "datetime":
		return "TIMESTAMP WITHOUT TIME ZONE NOT NULL", ""
	case "int", "int8", "int16", "int32", "int64":
		fallthrough
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "integer DEFAULT NULL", ""
	case "bool":
		return "boolean NOT NULL", ""
	case "float32", "float64", "float":
		return "numeric NOT NULL", ""
	}
	return "", ""
}

func NewDBDriver(sqlDriver string) DBDriver {
	switch sqlDriver {
	case "mysql":
		return mysqlDriver{}
	case "postgres":
		return postgresqlDriver{}
	default:
		beeLogger.Log.Fatal("Driver not supported")
		return nil
	}
}

// generateMigration generates migration file template for database schema update.
// The generated file template consists of an up() method for updating schema and
// a down() method for reverting the update.
func (c *Container) renderMigration(modelName string, content ModelsContent) (err error) {
	switch c.Option.SourceGen {
	case "text":
		c.textRenderMigration(modelName, content)
		return
	case "database":
		return
	}
	err = errors.New("not support source gen, source gen is " + c.Option.SourceGen)
	return
}

func (c *Container) textRenderMigration(mname string, content ModelsContent) {
	upsql := ""
	downsql := ""
	if len(content.Schema) != 0 {
		dbMigrator := NewDBDriver(c.Option.Driver)
		upsql = dbMigrator.GenerateCreateUp(mname, content.Schema)
		downsql = dbMigrator.GenerateCreateDown(mname)
	}
	today := time.Now().Format(MDateFormat)
	render := NewRenderGo("database", "migrations/"+mname, c.Option)
	render.SetContext("UpSQL", upsql)
	render.SetContext("StructName", utils.CamelCase(mname)+"_"+today)
	render.SetContext("DownSQL", downsql)
	render.SetContext("ddlSpec", "")
	render.SetContext("CurrTime", today)

	render.Exec("migrations/migrations.go.tmpl")
	migrate.MigrateUpdate(c.Option.BeegoPath, c.Option.Driver, c.Option.Dsn, "")
	return
}
