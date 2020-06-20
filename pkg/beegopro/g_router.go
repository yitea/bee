package beegopro

import (
	"database/sql"
	"errors"
	beeLogger "github.com/beego/bee/logger"
)

func (c *Container) renderRouter(modelName string, content ModelsContent) (err error) {
	switch content.SourceGen {
	case "text":
		c.textRenderRouters(modelName, content)
		return
	case "database":
		c.databaseRenderRouters(modelName, content)
		return
	}
	err = errors.New("not support source gen, source gen is " + content.SourceGen)
	return
}

func (c *Container) textRenderRouters(cname string, content ModelsContent) {
	render := NewRenderGo("routers", cname, c.Option)
	render.Exec("router.go.tmpl")
	return
}

func (c *Container) databaseRenderRouters(cname string, content ModelsContent) {
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

	tb := getTableObject(cname, db, trans)
	if tb.Pk == "" {
		return
	}
	render := NewRenderGo("routers", cname, c.Option)
	render.Exec("router.go.tmpl")
	return
}
