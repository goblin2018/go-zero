package mgo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

const serviceContextFile = "app/internal/svc/service_context.go"

func UpdateServiceContext(modelName string) error {
	modelType := fmt.Sprintf("*model.%s", modelName)
	// 解析 service.go 文件
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, serviceContextFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// 查找并更新 ServiceContext 结构体和 NewServiceContext 函数
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if x.Name.Name == "ServiceContext" {
				st, ok := x.Type.(*ast.StructType)
				if !ok {
					return false
				}

				// 添加新的模型到 ServiceContext 结构体
				addModelToStruct(st, modelName, modelType)
			}
		case *ast.FuncDecl:
			if x.Name.Name == "NewServiceContext" {
				// 添加新的模型到 NewServiceContext 函数
				addModelToFunc(x, modelName)
			}
		}
		return true
	})

	// 将更新后的内容写回 service.go 文件
	file, err := os.Create(serviceContextFile)
	if err != nil {
		return err
	}
	defer file.Close()

	err = printer.Fprint(file, fset, node)
	if err != nil {
		return err
	}

	return nil
}

// 向 ServiceContext 结构体添加新的模型
func addModelToStruct(st *ast.StructType, modelName, modelType string) {
	for _, field := range st.Fields.List {
		if field.Names[0].Name == modelName {
			// 模型已存在，不需要添加
			return
		}
	}

	// 添加新的模型
	st.Fields.List = append(st.Fields.List, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(modelName)},
		Type:  ast.NewIdent(modelType),
	})
}

// 向 NewServiceContext 函数添加新的模型初始化
func addModelToFunc(fn *ast.FuncDecl, modelName string) {
	// 查找 return 语句
	for _, stmt := range fn.Body.List {
		retStmt, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		// 查找 return &ServiceContext{...} 语句
		for _, expr := range retStmt.Results {
			ce, ok := expr.(*ast.UnaryExpr)
			if !ok {
				continue
			}

			if ident, ok := ce.X.(*ast.CompositeLit); ok {
				if structIdent, ok := ident.Type.(*ast.Ident); ok && structIdent.Name == "ServiceContext" {
					// 检查模型是否已经存在
					for _, elt := range ident.Elts {
						kve := elt.(*ast.KeyValueExpr)
						if kve.Key.(*ast.Ident).Name == modelName {
							// 模型已存在，不需要添加
							return
						}
					}

					// 添加新的模型初始化
					ident.Elts = append(ident.Elts, &ast.KeyValueExpr{
						Key:   ast.NewIdent("\n\t\t" + modelName),
						Value: ast.NewIdent("model.New" + modelName + "(),\n"),
					})
				}
			}
		}
	}
}
