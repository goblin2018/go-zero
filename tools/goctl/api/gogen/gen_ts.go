package gogen

import (
	_ "embed"
	"reflect"

	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	apiutil "github.com/zeromicro/go-zero/tools/goctl/api/util"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/util"
)

//go:embed ts.tpl
var tsTemplate string

// todo dir用参数传入
const tsDir = "types"

// 生成model的时候 只需要生成相关model 和 子model

// BuildTypes gen types to string
func BuildTsInterface(types []spec.Type) (string, error) {
	var builder strings.Builder
	for _, tp := range types {
		st, ok := tp.(spec.DefineStruct)
		if !ok {
			continue
		}

		if len(st.Docs) == 0 {
			continue
		}

		builder.Write([]byte("\n"))

		d := st.Docs[0]
		d = strings.TrimPrefix(d, "// ")
		d = strings.TrimSpace(d)

		fmt.Println("d:", d)
		if (d != "base") && (d != "child") {
			continue
		}
		if err := writeTsInterface(&builder, st); err != nil {
			return "", apiutil.WrapErr(err, "Type "+tp.Name()+" generate error")
		}
	}

	return builder.String(), nil
}

func genTs(dir string, cfg *config.Config, api *spec.ApiSpec, name string) error {

	val, err := BuildTsInterface(api.Types)
	if err != nil {
		return err
	}

	typeFilename := name + ".ts"
	fullpath := path.Join(tsDir, typeFilename)
	fmt.Println("typeFilename:", fullpath)
	os.Remove(fullpath)

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          tsDir,
		filename:        typeFilename,
		templateName:    "tsTemplate",
		category:        category,
		templateFile:    "ts.tpl",
		builtinTemplate: tsTemplate,
		data: map[string]any{
			"types":        val,
			"containsTime": false,
		},
	})
}

func writeTsInterface(writer io.Writer, st spec.DefineStruct) error {

	fmt.Fprintf(writer, "export interface %s {\n", util.Title(st.Name()))
	fmt.Fprint(writer, "\tkey?: string\n")
	fmt.Fprint(writer, "\tid?: string\n")
	for _, member := range st.Members {
		if err := writeItem(writer, member.Name, member.Tag, member.GetComment(), member.Type, 1); err != nil {
			return err
		}
	}
	fmt.Fprintf(writer, "}")
	return nil
}
func writeItem(writer io.Writer, name, tag, comment string, tp spec.Type, indent int) error {
	// id 自动写入
	if name == "Id" {
		return nil
	}
	var err error
	// 对tag 进行处理，optinal 修改为omitempty
	// 增加bson 注释 其中 id 改为 _id
	rt := reflect.StructTag(strings.Trim(tag, "`"))

	// 处理json tag
	jt := rt.Get("json")
	if jt == "" {
		return nil
	}
	jts := strings.Split(jt, ",")
	name = jts[0]

	tname := castToTs(tp.Name())

	apiutil.WriteIndent(writer, indent)
	if len(comment) > 0 {
		comment = strings.TrimPrefix(comment, "//")
		comment = "//" + comment
		_, err = fmt.Fprintf(writer, "%s?: %s %s\n", name, tname, comment)
	} else {
		_, err = fmt.Fprintf(writer, "%s?: %s\n", name, tname)
	}

	return err
}

func castToTs(goType string) string {
	t := strings.TrimPrefix(goType, "Simple")
	switch t {
	case "int", "int32", "int64", "uint", "uint32", "uint64", "float32", "float64":
		return "number"
	case "string":
		return "string"
	case "bool":
		return "boolean"
	default:
		if strings.HasPrefix(t, "*") {
			return castToTs(t[1:])
		} else if strings.HasPrefix(t, "[]") {
			return castToTs(t[2:]) + "[]"
		} else if strings.HasPrefix(t, "map[") {
			end := strings.Index(t, "]")
			keyType := castToTs(t[4:end])
			valueType := castToTs(t[end+1:])
			return "{ [key: " + keyType + "]: " + valueType + " }"
		}
		return t
	}
}
