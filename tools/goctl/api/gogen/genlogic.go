package gogen

import (
	"bytes"
	_ "embed"
	"fmt"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/zeromicro/go-zero/tools/goctl/api/parser/g4/gen/api"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/vars"
)

//go:embed logic.tpl
var logicTemplate string

func genLogic(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	for _, g := range api.Service.Groups {
		for _, r := range g.Routes {
			err := genLogicByRoute(dir, rootPkg, cfg, g, r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func genLogicByRoute(dir, rootPkg string, cfg *config.Config, group spec.Group, route spec.Route) error {
	logic := getLogicName(route)
	goFile, err := format.FileNamingFormat(cfg.NamingFormat, logic)
	if err != nil {
		return err
	}

	imports := genLogicImports(route, rootPkg)
	var responseString string
	var returnString string
	var requestString string
	if len(route.ResponseTypeName()) > 0 {
		resp := responseGoTypeName(route, typesPacket)
		responseString = "(resp " + resp + ", err error)"
		returnString = "return"
	} else {
		responseString = "(err error)"
		returnString = "return"
	}
	if len(route.RequestTypeName()) > 0 {
		requestString = "req *" + requestGoTypeName(route, typesPacket)
	}
	funcName := strings.TrimSuffix(logic, "Logic")
	bodyStr := ""

	// 获取请求类型名称 去除前缀和后缀
	requestType := route.RequestType.Name()
	requestType = strings.TrimPrefix(requestType, "List")
	requestType = strings.TrimSuffix(requestType, "Req")
	requestType = strings.TrimPrefix(requestType, "Del")

	// 增加增删改查的基础逻辑
	switch funcName {
	case "add":
		bodyStr = genBody(requestType, addTemplate)
	case "update":
		bodyStr = genBody(requestType, updateTemplate)
	case "del":
		bodyStr = genBody(requestType, delTemplate)
	case "list":
		bodyStr = genBody(requestType, listTemplate)
	}

	subDir := getLogicFolderPath(group, route)
	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          subDir,
		filename:        goFile + ".go",
		templateName:    "logicTemplate",
		category:        category,
		templateFile:    logicTemplateFile,
		builtinTemplate: logicTemplate,
		data: map[string]string{
			"pkgName":      subDir[strings.LastIndex(subDir, "/")+1:],
			"imports":      imports,
			"logic":        strings.Title(logic),
			"function":     strings.Title(strings.TrimSuffix(logic, "Logic")),
			"body":         bodyStr,
			"responseType": responseString,
			"returnString": returnString,
			"request":      requestString,
		},
	})
}

func getLogicFolderPath(group spec.Group, route spec.Route) string {
	folder := route.GetAnnotation(groupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(groupProperty)
		if len(folder) == 0 {
			return logicDir
		}
	}
	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")
	return path.Join(logicDir, folder)
}

func genLogicImports(route spec.Route, parentPkg string) string {
	var imports []string
	imports = append(imports, `"context"`+"\n")
	imports = append(imports, fmt.Sprintf("\"%s\"", pathx.JoinPackages(parentPkg, contextDir)))
	if shallImportTypesPackage(route) {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", pathx.JoinPackages(parentPkg, typesDir)))
	}
	imports = append(imports, fmt.Sprintf("\"%s/core/logx\"", vars.ProjectOpenSourceURL))
	return strings.Join(imports, "\n\t")
}

func onlyPrimitiveTypes(val string) bool {
	fields := strings.FieldsFunc(val, func(r rune) bool {
		return r == '[' || r == ']' || r == ' '
	})

	for _, field := range fields {
		if field == "map" {
			continue
		}
		// ignore array dimension number, like [5]int
		if _, err := strconv.Atoi(field); err == nil {
			continue
		}
		if !api.IsBasicType(field) {
			return false
		}
	}

	return true
}

func shallImportTypesPackage(route spec.Route) bool {
	if len(route.RequestTypeName()) > 0 {
		return true
	}

	respTypeName := route.ResponseTypeName()
	if len(respTypeName) == 0 {
		return false
	}

	if onlyPrimitiveTypes(respTypeName) {
		return false
	}

	return true
}

const addTemplate = `
md := new(model.{{.Name}})
copier.Copy(md, req)
err = l.{{.Name}}Model.Create(md)
`

const updateTemplate = `
md := new(model.{{.Name}})
copier.Copy(md, req)
err = l.{{.Name}}Model.Update(md)
`

const delTemplate = `
err = l.{{.Name}}Model.Delete(req.Id)
`

const listTemplate = `
resp = new(types.List{{.Name}}Resp)
	items, _ := l.{{.Name}}Model.List(l.ctx, req.Page, req.Size, req.All)
	for _, it := range items {
		td := new(types.{{.Name}})
		copier.Copy(td, it)
		resp.Items = append(resp.Items, td)
	}

	if !req.All {
		resp.Total, _ = l.{{.Name}}Model.Count(l.ctx)
	}`

// 添加逻辑
func genBody(name string, t string) string {
	tmpl, _ := template.New("abc").Parse(t)

	var buf bytes.Buffer
	tmpl.Execute(&buf, map[string]interface{}{
		"Name": name,
	})

	return buf.String()
}
