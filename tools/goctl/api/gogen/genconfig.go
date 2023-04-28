package gogen

import (
	_ "embed"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
)

const (
	configFile = "config"
)

//go:embed config.tpl
var configTemplate string

func genConfig(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	filename, err := format.FileNamingFormat(cfg.NamingFormat, configFile)
	if err != nil {
		return err
	}

	// config 文件是固定的，不需要根据配置文件的命名规则来生成
	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          configDir,
		filename:        filename + ".go",
		templateName:    "configTemplate",
		category:        category,
		templateFile:    configTemplateFile,
		builtinTemplate: configTemplate,
		data:            map[string]string{},
	})
}
