package apigen

import (
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

// Todo 自定义path和group 强制删除

//go:embed api.tpl
var apiTemplate string

var (
	// VarStringOutput describes the output.
	VarStringOutput string
	// VarStringHome describes the goctl home.
	VarStringHome string
	// VarStringRemote describes the remote git repository.
	VarStringRemote string
	// VarStringBranch describes the git branch.
	VarStringBranch string
)

// CreateApiTemplate create api template file
func CreateApiTemplate(_ *cobra.Command, _ []string) error {
	apiFile := VarStringOutput
	if len(apiFile) == 0 {
		return errors.New("missing -o")
	}

	fp, err := pathx.CreateIfNotExist(apiFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	// 处理远程和本地模板文件问题
	if len(VarStringRemote) > 0 {
		repo, _ := util.CloneIntoGitHome(VarStringRemote, VarStringBranch)
		if len(repo) > 0 {
			VarStringHome = repo
		}
	}

	if len(VarStringHome) > 0 {
		pathx.RegisterGoctlHome(VarStringHome)
	}

	text, err := pathx.LoadTemplate(category, apiTemplateFile, apiTemplate)
	if err != nil {
		return err
	}

	path, name := GetName(apiFile)

	t := template.Must(template.New("etcTemplate").Parse(text))
	if err := t.Execute(fp, map[string]string{
		"name":  name,
		"group": path,
		"path":  path,
	}); err != nil {
		return err
	}

	fmt.Println(color.Green.Render("Done."))
	return nil
}

func GetName(apiPath string) (string, string) {
	parts := strings.Split(apiPath, "/")

	// 获取 "test_me"
	filename := parts[len(parts)-1]
	filename = strings.TrimSuffix(filename, ".api")

	// 将 "test_me" 转换为 "TestMe"
	camelCase := toCamelCase(filename)
	return filename, camelCase
}

func toCamelCase(str string) string {
	words := strings.Split(str, "_")
	camelCase := ""

	for _, word := range words {
		camelCase += strings.ToUpper(word[:1])
		if len(word) > 1 {
			camelCase += word[1:]
		}
	}

	return camelCase
}
