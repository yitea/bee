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
	"database/sql"
	"errors"
	"fmt"
	beeLogger "github.com/beego/bee/logger"
	"github.com/beego/bee/utils"
	"os"
	"strings"
)

func (c *Container) renderModel(modelName string, content ModelsContent) (err error) {
	switch content.SourceGen {
	case "text":
		c.TextRenderModel(modelName, content)
	case "database":
		c.databaseRenderModel(modelName, content)
	default:
		err = errors.New("not support source gen, source gen is " + content.SourceGen)
		return
	}
	c.ModelOnce.Do(func() {
		render := NewRenderGo("models", "bee_default_model", c.Option)
		if utils.IsExist(render.TmplPath + "/" + BeeDefaultModelTmpl) {
			render.Exec(BeeDefaultModelTmpl)
		}
	})

	return
}

func (c *Container) TextRenderModel(mname string, content ModelsContent) {
	render := NewRenderGo("models", mname, c.Option)

	modelSchemas, hasTime := initModel(render.Name, content.Schema)
	fmt.Println("modelSchemas------>", modelSchemas)
	importMaps := make(map[string]struct{}, 0)
	if hasTime {
		importMaps["time"] = struct{}{}
	}

	camelPrimaryKey := ""
	for _, value := range modelSchemas {
		if value.ColumnKey == "PRI" {
			camelPrimaryKey = utils.CamelString(value.Name)
		}
	}

	render.SetContext("imports", importMaps)
	render.SetContext("modelSchemas", modelSchemas)
	render.SetContext("primaryKey", camelPrimaryKey)
	render.SetContext("tableName", utils.SnakeString(render.Name))

	render.Exec("model.go.tmpl")
}

func (c *Container) databaseRenderModel(mname string, content ModelsContent) {
	// todo uniform sql open
	db, err := sql.Open(c.Option.Driver, c.Option.Dsn)
	if err != nil {
		beeLogger.Log.Fatalf("Could not connect to '%s' database using '%s': %s", c.Option.Driver, c.Option.Dsn, err)
		return
	}

	defer db.Close()

	trans, ok := dbDriver[c.Option.Driver]
	if !ok {
		beeLogger.Log.Fatalf("Generating app code from '%s' database is not supported yet.", c.Option.Driver)
		return
	}

	tb := getTableObject(mname, db, trans)

	render := NewRenderGo("models", utils.CamelCase(tb.Name), c.Option)

	render.SetContext("modelStruct", tb.String())
	render.SetContext("tableName", utils.SnakeString(render.Name))

	// If table contains time field, import time.Time package
	if tb.ImportTimePkg {
		render.SetContext("timePkg", "\"time\"\n")
		render.SetContext("importTimePkg", "import \"time\"\n")
	}

	render.Exec("model.go.tmpl")
}

func initModel(tableName string, schema []Schema) (output []ModelSchema, hasTime bool) {
	output = make([]ModelSchema, 0)
	for i, v := range schema {
		columnKey := ""
		if i == 0 && v.Type != "autopk" && strings.ToLower(v.Name) != "id" {
			kt, ormTag, isImportTime := getModelType("auto")
			if isImportTime {
				hasTime = true
			}
			output = append(output, ModelSchema{
				Name:      upperFirst("id"),
				CamelName: upperFirst("id"),
				Type:      v.Type,
				ColumnKey: "PRI",
				Comment:   v.Comment,
				GoType:    kt,
				GoJsonTag: lowerFirst("id"),
				OrmTag:    ormTag,
			})
		}
		kt, ormTag, isImportTime := getModelType(v.Type)
		if isImportTime {
			hasTime = true
		}

		if v.Type == "autopk" {
			columnKey = "PRI"
		}
		output = append(output, ModelSchema{
			Name:      upperFirst(v.Name),
			CamelName: upperFirst(v.Name),
			Type:      v.Type,
			ColumnKey: columnKey,
			Comment:   v.Comment,
			GoType:    kt,
			GoJsonTag: lowerFirst(v.Name),
			OrmTag:    ormTag,
		})
	}
	return
}

func getModelType(ktype string) (kt, tag string, hasTime bool) {
	kv := strings.SplitN(ktype, ":", 2)
	switch kv[0] {
	case "string":
		if len(kv) == 2 {
			return "string", "size(" + kv[1] + ")\"`", false
		}
		return "string", "size(128)", false
	case "text":
		return "string", "type(longtext)", false
	case "auto":
		return "int", "auto", false
	case "autopk":
		return "int", "auto", false
	case "pk":
		return "int", "pk", false
	case "datetime":
		return "time.Time", "type(datetime)", true
	case "int", "int8", "int16", "int32", "int64":
		fallthrough
	case "uint", "uint8", "uint16", "uint32", "uint64":
		fallthrough
	case "bool":
		fallthrough
	case "float32", "float64":
		return kv[0], "", false
	case "float":
		return "float64", "", false
	}
	return "", "", false
}

// createPath 调用os.MkdirAll递归创建文件夹
func createPath(filePath string) error {
	if !utils.IsExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		return err
	}
	return nil
}
