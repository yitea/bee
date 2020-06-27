package beegopro

import (
	"errors"
	"fmt"
	beeLogger "github.com/beego/bee/logger"
	"github.com/beego/bee/utils"
	"github.com/flosch/pongo2"
	"github.com/smartwalle/pongo2render"
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

// render
type RenderGo struct {
	Context pongo2.Context
	Option  Option
	*pongo2render.Render
	Name        string
	PackageName string
	FlushFile   string
	PkgPath     string
	TmplPath    string
}

func NewRenderGo(packageName string, name string, option Option) *RenderGo {
	language := "go"
	p, f := path.Split(name)
	title := strings.Title(f)

	obj := &RenderGo{
		Context:     make(pongo2.Context, 0),
		Option:      option,
		Name:        title,
		PackageName: packageName,
		TmplPath:    option.GitLocalPath + "/" + option.ProType + "/" + option.ProVersion + "/" + language + "/" + packageName,
	}
	// render path
	obj.Render = pongo2render.NewRender(obj.TmplPath)

	//if p != "" {
	//	i := strings.LastIndex(p[:len(p)-1], "/")
	//	packageName = p[i+1 : len(p)-1]
	//}

	beeLogger.Log.Infof("Using '%s' as name from %s", title, obj.PackageName)
	beeLogger.Log.Infof("Using '%s' as package name from %s", packageName, obj.PackageName)

	fp := path.Join(obj.Option.BeegoPath, packageName, p)
	err := createPath(fp)
	if err != nil {
		beeLogger.Log.Fatalf("Could not create the controllers directory: %s", err)
	}

	obj.FlushFile = path.Join(fp, strings.ToLower(title)+".go")
	obj.PkgPath = getPackagePath(obj.Option.BeegoPath)

	obj.Context["packageName"] = obj.PackageName
	obj.Context["name"] = obj.Name
	obj.Context["pkgPath"] = obj.PkgPath
	obj.Context["apiPrefix"] = obj.Option.ApiPrefix
	return obj
}

func (r *RenderGo) SetContext(key string, value interface{}) {
	r.Context[key] = value
}

func (r *RenderGo) Exec(name string) {
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
	beeLogger.Log.Infof("create file '%s' from %s", r.FlushFile, r.PackageName)
}

// write 写bytes到文件
func (c *RenderGo) write(filename string, buf string) (err error) {
	// 不允许覆盖
	if utils.IsExist(filename) && !isNeedGoOverwrite(filename) {
		return
	}

	//overwrite := false
	//if utils.IsExist(filename) && isNeedGoOverwrite(filename) {
	//	overwrite = true
	//}
	//
	//if !c.Option.Overwrite && utils.IsExist(filename) {
	//	err = errors.New("file is exist, path is " + filename)
	//	return
	//}

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

	output := []byte(buf)

	if c.Option.Format {
		// 格式化代码
		var bts []byte
		bts, err = format.Source([]byte(buf))
		if err != nil {
			err = errors.New("format buf error " + err.Error())
			return
		}
		output = bts
	}

	err = ioutil.WriteFile(filename, output, 0644)
	if err != nil {
		err = errors.New("write write file " + err.Error())
		return
	}
	return
}
