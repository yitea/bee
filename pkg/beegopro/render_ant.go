package beegopro

import (
	"errors"
	"fmt"
	beeLogger "github.com/beego/bee/logger"
	"github.com/beego/bee/utils"
	"github.com/flosch/pongo2"
	"github.com/smartwalle/pongo2render"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

// render
type RenderAnt struct {
	Context pongo2.Context
	Option  Option
	*pongo2render.Render
	Name      string
	TypeName  string
	FlushFile string
	PkgPath   string
}

// type name list, create update, name biz name
func NewRenderAnt(typeName string, name string, option Option) *RenderAnt {
	language := "ant"
	p, f := path.Split(name)
	title := strings.Title(f)

	obj := &RenderAnt{
		Context:  make(pongo2.Context, 0),
		Option:   option,
		Name:     title,
		TypeName: typeName,
	}
	// render path
	obj.Render = pongo2render.NewRender(option.GitPath + "/" + option.ProType + "/" + option.ProVersion + "/" + language + "/")

	if p != "" {
		i := strings.LastIndex(p[:len(p)-1], "/")
		typeName = p[i+1 : len(p)-1]
	}

	beeLogger.Log.Infof("Using '%s' as name from %s", title, obj.TypeName)
	beeLogger.Log.Infof("Using '%s' as package name from %s", typeName, obj.TypeName)

	fp := path.Join(obj.Option.AntDesignPath, p, name)
	err := createPath(fp)
	if err != nil {
		beeLogger.Log.Fatalf("Could not create the controllers directory: %s", err)
	}

	obj.FlushFile = path.Join(fp, strings.ToLower(typeName)+".tsx")
	obj.PkgPath = getPackagePath(obj.Option.BeegoPath)

	obj.Context["typeName"] = obj.TypeName
	obj.Context["name"] = obj.Name
	obj.Context["pkgPath"] = obj.PkgPath
	obj.Context["apiPrefix"] = obj.Option.ApiPrefix
	return obj
}

func (r *RenderAnt) SetContext(key string, value interface{}) {
	r.Context[key] = value
}

func (r *RenderAnt) Exec(name string) {
	var (
		buf string
		err error
	)
	buf, err = r.Render.Template(name).Execute(r.Context)
	if err != nil {
		beeLogger.Log.Fatalf("Could not create the %s render tmpl: %s", name, err)
		return
	}
	err = r.write(r.FlushFile, buf)
	if err != nil {
		beeLogger.Log.Fatalf("Could not create file: %s", err)
		return
	}
	beeLogger.Log.Infof("create file '%s' from %s", r.FlushFile, r.TypeName)
}

// write 写bytes到文件
func (c *RenderAnt) write(filename string, buf string) (err error) {
	// 不允许覆盖
	if utils.IsExist(filename) && !isNeedOverwrite(filename) {
		return
	}

	filePath := path.Dir(filename)
	err = createPath(filePath)
	if err != nil {
		err = errors.New("write create path " + err.Error())
		return
	}
	filePathBak := filePath + "/bak"
	err = createPath(filePathBak)
	if err != nil {
		err = errors.New("write create path bak " + err.Error())
		return
	}

	name := path.Base(filename)

	if utils.IsExist(filename) {
		bakName := fmt.Sprintf("%s/%s.%s.bak", filePathBak, name, time.Now().Format("2006.01.02.15.04.05"))
		beeLogger.Log.Infof("bak file '%s'", bakName)
		if err := os.Rename(filename, bakName); err != nil {
			err = errors.New("file is bak error, path is " + bakName)
			return err
		}
	}

	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		err = errors.New("write create file " + err.Error())
		return
	}

	err = ioutil.WriteFile(filename, []byte(buf), 0644)
	if err != nil {
		err = errors.New("write write file " + err.Error())
		return
	}
	return
}
