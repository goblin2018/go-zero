package mgo

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"text/template"
)

//go:embed model.tpl
var modelTemplate string

const dstDir = "app/model/"

// model 结构体
type TemplateData struct {
	StructName     string
	TableName      string
	Unis           []KeyInfo // 唯一索引
	UnisWC         string    // 中间使用:冒号
	UnisWD         string    // 中间使用,逗号
	UnisWDWQ       string    // 中间使用,逗号 加引号
	UnisWAnd       string    // 中间使用and符号
	UnisWType      string    // 唯一索引带类型
	UnisPD         string    // 前面带data.
	UnisWDs        string    // 中间使用,逗号
	UnisWTWT       string    // tag 引号 tag
	FileName       string
	UseList        bool   // 是否使用列表
	ListFilter     string // 获取列表的过滤条件
	ListStructName string // 获取列表的结构体名称
	UsePage        bool   // 是否使用分页
	UseCountFilter bool   // 是否使用分页的过滤条件
}

// 索引信息
type KeyInfo struct {
	Field string // 字段名
	Type  string // 类型名称
	Tag   string // 索引名称
}

func GenMongo(name string) (err error) {
	filePath := fmt.Sprintf("%s%s.go", dstDir, name)
	// 解析结构体文件
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return
	}
	// 查找第一个结构体定义
	sts := findFirstAndListStruct(node)

	// 解析模板文件
	tmpl, err := template.New("mgoTemplate").Parse(modelTemplate)
	if err != nil {
		fmt.Println("parse template failed ", err)
		return
	}

	// 模板数据
	td := getTemplateData(sts, name)

	// 生成代码并保存到xxx.gen.go文件中
	generateAndSaveCode(td, tmpl)

	UpdateServiceContext(td.StructName + "Model")

	return
}

// 获取模板数据
func getTemplateData(sts []StructInfo, filename string) TemplateData {

	fst := sts[0]

	td := TemplateData{
		StructName: fst.Name,
		TableName:  ToSnakeCase(fst.Name),
		FileName:   filename,
		Unis:       make([]KeyInfo, 0),
	}

	// 遍历结构体的字段
	// tag字段
	var tags []string
	var fields []string // 字段
	var ts []string     // 类型
	for _, field := range fst.AstType.Fields.List {
		tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
		// keyTag := tag.Get("key")
		uniTag := tag.Get("uni")
		bsonTag := tag.Get("bson")
		if uniTag != "" {
			tags = append(tags, bsonTag)
			fields = append(fields, field.Names[0].Name)
			ts = append(ts, field.Type.(*ast.Ident).Name)
			td.Unis = append(td.Unis, KeyInfo{Field: field.Names[0].Name, Type: field.Type.(*ast.Ident).Name, Tag: bsonTag})
		}

	}

	td.UnisWC = strings.Join(tags, ":")
	td.UnisWAnd = strings.Join(fields, "and")
	td.UnisWD = strings.Join(tags, ", ")
	for i, v := range tags {
		td.UnisWType += fmt.Sprintf("%s %s", v, ts[i])
		td.UnisWDWQ += fmt.Sprintf("\"%s\"", v)
		td.UnisPD += fmt.Sprintf("data.%s", fields[i])
		td.UnisWTWT += fmt.Sprintf("\"%s\": %s", v, v)
		if i != len(tags)-1 {
			td.UnisWType += ", "
			td.UnisWDWQ += ", "
			td.UnisPD += ", "
			td.UnisWTWT += ", "
		}
	}

	// 处理列表数据
	var filter string
	if len(sts) > 1 {
		td.UseList = true
		listStruct := sts[1]
		td.ListStructName = listStruct.Name
		filter = "bson.M{"
		for _, field := range listStruct.AstType.Fields.List {
			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			jsonTag := tag.Get("json")
			fmt.Println("jsonTag: ", jsonTag)
			sts := strings.Split(jsonTag, ",")
			jsonTag = sts[0]
			if jsonTag == "page" {
				td.UsePage = true
				continue
			}
			if jsonTag == "size" || jsonTag == "all" {

				continue
			}
			if jsonTag != "" {
				filter += fmt.Sprintf("\"%s\": req.%s, ", jsonTag, field.Names[0].Name)
				td.UseCountFilter = true
			}
		}
		filter += "}"
	}
	td.ListFilter = filter
	fmt.Println("filter: ", filter)
	return td
}

type StructInfo struct {
	AstType *ast.StructType
	Name    string
}

// 查找第一个结构体
func findFirstAndListStruct(node *ast.File) (items []StructInfo) {

	hasFirst := false
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			r, ok := typeSpec.Type.(*ast.StructType)
			if ok && !hasFirst {
				items = append(items, StructInfo{AstType: r, Name: typeSpec.Name.Name})
				hasFirst = true
			}

			if hasFirst {
				if strings.HasPrefix(typeSpec.Name.Name, "List") {
					items = append(items, StructInfo{AstType: r, Name: typeSpec.Name.Name})
					break
				}
			}
		}
	}
	return
}

// 生成代码并保存到xxx.gen.go文件中
func generateAndSaveCode(td TemplateData, tmpl *template.Template) {

	// 执行模板，生成代码
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, td)
	if err != nil {
		fmt.Println(err)
		return
	}
	generatedCode := buf.String()

	// 生成的文件名
	genFileName := fmt.Sprintf("%s%s.gen.go", dstDir, td.FileName)

	// 如果文件已存在，则删除
	if _, err := os.Stat(genFileName); err == nil {
		err = os.Remove(genFileName)
		if err != nil {
			fmt.Println("Error removing existing file:", err)
			return
		}
	}

	// 将生成的代码保存到xxx.gen.go文件中
	err = ioutil.WriteFile(genFileName, []byte(generatedCode), os.ModePerm)
	if err != nil {
		fmt.Println("Error writing file:", err)
	}
}
