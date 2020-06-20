package beegopro

import (
	beeLogger "github.com/beego/bee/logger"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
)

// todo
func isNeedGoOverwrite(fileName string) (flag bool) {
	// 判断注释上是否有关键字
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		beeLogger.Log.Fatalf("Error while parsing file: %s", err)
		return
	}

	overwrite := ""

	// Analyse comments
	if f.Comments != nil {
		for _, c := range f.Comments {
			for _, s := range strings.Split(c.Text(), "\n") {
				if strings.HasPrefix(s, "@BeeOverwrite") {
					overwrite = strings.TrimSpace(s[len("@BeeOverwrite"):])
				}
			}
		}
	}

	if strings.ToLower(overwrite) == "yes" {
		flag = true
		return
	}
	return
}

func isNeedOverwrite(fileName string) (flag bool) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()
	overwrite := ""
	var contentByte []byte
	contentByte, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}
	for _, s := range strings.Split(string(contentByte), "\n") {
		s = strings.TrimSpace(strings.TrimPrefix(s, "//"))
		if strings.HasPrefix(s, "@BeeOverwrite") {
			overwrite = strings.TrimSpace(s[len("@BeeOverwrite"):])
		}
	}
	if strings.ToLower(overwrite) == "yes" {
		flag = true
		return
	}
	return
}
