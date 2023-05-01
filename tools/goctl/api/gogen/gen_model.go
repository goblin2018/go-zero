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

//go:embed model.tpl
var modelTemplate string

const modelDir = "model"

// 生成model的时候 只需要生成相关model 和 子model

// BuildTypes gen types to string
func BuidModel(types []spec.Type) (string, error) {
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
		if err := writeT(&builder, st); err != nil {
			return "", apiutil.WrapErr(err, "Type "+tp.Name()+" generate error")
		}
	}

	return builder.String(), nil
}

// todo 分开多个文件生成
func genModel(dir string, cfg *config.Config, api *spec.ApiSpec, name string) error {
	fmt.Println("genModel")

	val, err := BuidModel(api.Types)
	if err != nil {
		return err
	}

	typeFilename := name + ".go"
	filename := path.Join(dir, modelDir, typeFilename)
	fmt.Println("typeFilename:", filename)
	os.Remove(filename)

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          modelDir,
		filename:        typeFilename,
		templateName:    "modelTemplate",
		category:        category,
		templateFile:    "model.tpl",
		builtinTemplate: modelTemplate,
		data: map[string]any{
			"types":        val,
			"containsTime": false,
		},
	})
}

func writeT(writer io.Writer, st spec.DefineStruct) error {

	fmt.Fprintf(writer, "type %s struct {\n", util.Title(st.Name()))
	fmt.Fprint(writer, "Id string `json:\"id\" bson:\"_id,omitempty\"`\n")
	fmt.Fprint(writer, "CreateAt time.Time `json:\"createAt\" bson:\"createAt\"`\n")
	fmt.Fprint(writer, "UpdateAt time.Time `json:\"updateAt\" bson:\"updateAt\"`\n")
	for _, member := range st.Members {
		if member.IsInline {
			if _, err := fmt.Fprintf(writer, "%s\n", strings.Title(member.Type.Name())); err != nil {
				return err
			}

			continue
		}

		if err := writeModelP(writer, member.Name, member.Tag, member.GetComment(), member.Type, 1); err != nil {
			return err
		}
	}
	fmt.Fprintf(writer, "}")
	return nil
}
func writeModelP(writer io.Writer, name, tag, comment string, tp spec.Type, indent int) error {
	// id 自动写入
	if name == "Id" {
		return nil
	}
	var err error
	// 对tag 进行处理，optinal 修改为omitempty
	// 增加bson 注释 其中 id 改为 _id
	rt := reflect.StructTag(strings.Trim(tag, "`"))
	// 这时候不写入结构体
	if rt.Get("apionly") == "true" {
		return nil
	}
	jt := rt.Get("json")
	jts := strings.Split(jt, ",")
	if len(jts) > 1 {
		if jts[1] == "optional" {
			jt = jts[0] + ",omitempty"
		}
	}

	fmt.Println("get json tag: ", jt)

	ntag := fmt.Sprintf("`json:\"%s\" bson:\"%s\"", jt, jt)
	if rt.Get("key") != "" {
		ntag = fmt.Sprintf("%s key:\"%s\"", ntag, rt.Get("key"))
	}
	if rt.Get("uni") != "" {
		ntag = fmt.Sprintf("%s uni:\"%s\"", ntag, rt.Get("uni"))
	}
	ntag = fmt.Sprintf("%s`", ntag)

	apiutil.WriteIndent(writer, indent)
	if len(comment) > 0 {
		comment = strings.TrimPrefix(comment, "//")
		comment = "//" + comment
		_, err = fmt.Fprintf(writer, "%s %s %s %s\n", strings.Title(name), tp.Name(), ntag, comment)
	} else {
		_, err = fmt.Fprintf(writer, "%s %s %s\n", strings.Title(name), tp.Name(), ntag)
	}

	return err
}
