package beegopro

import (
	"database/sql"
	"errors"
	beeLogger "github.com/beego/bee/logger"
	"github.com/beego/bee/utils"
)

func (c *Container) renderAntList(modelName string, content ModelsContent) (err error) {
	switch c.Option.SourceGen {
	case "text":
		c.textRenderAntList(modelName, content)
		return
	case "database":
		c.databaseRenderAntList(modelName, content)
		return
	}
	err = errors.New("not support source gen, source gen is " + c.Option.SourceGen)
	return
}

func (c *Container) textRenderAntList(mname string, content ModelsContent) {
	render := NewRenderAnt("list", mname, c.Option)

	modelSchemas, _ := initModel(render.Name, content.Schema)
	camelPrimaryKey := initPrimaryKey(modelSchemas)
	render.SetContext("primaryKey", camelPrimaryKey)
	render.SetContext("lowerFirstPrimaryKey", lowerFirst(camelPrimaryKey))

	render.SetContext("modelSchemas", modelSchemas)
	render.SetContext("apiUrl", c.Option.ApiPrefix+"/"+mname)
	render.SetContext("pageCreate", "/"+mname+"/create")
	render.SetContext("pageUpdate", "/"+mname+"/update")
	render.SetContext("pageInfo", "/"+mname+"/info")
	render.SetContext("tableName", utils.SnakeString(render.Name))

	render.Exec("list.tsx.tmpl")
}

func (c *Container) databaseRenderAntList(mname string, content ModelsContent) {
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

	render := NewRenderAnt("list", mname, c.Option)

	columns := make([]AntColumn, 0)
	for _, column := range tb.Columns {
		title := column.Tag.Comment
		if title == "" {
			title = column.Name
		}
		columns = append(columns, AntColumn{
			Title: title,
			Key:   column.Name,
		})
	}

	render.SetContext("columns", columns)
	render.SetContext("apiList", c.Option.ApiPrefix+"/"+mname)
	render.SetContext("pageCreate", "/"+mname+"/create")
	render.SetContext("tableName", utils.SnakeString(render.Name))

	render.Exec("list.tsx.tmpl")
}

type AntColumn struct {
	Title string `json:"title"`
	Key   string `json:"key"`
}
