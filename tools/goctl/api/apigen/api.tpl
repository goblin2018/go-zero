syntax = "v1"

type {{.name}} {
	Id string `json:"id,optional"`
}

// 删除对象请求
type Del{{.name}}Req {
	Id string `json:"id"`
}

// 获取列表请求
type (
  List{{.name}}Req {
		Page int64 `json:"page,optional"`
		Size int64 `json:"size,optional"`
		All bool `json:"all,optional"`
	}

	List{{.name}}Resp {
		Items []*{{.name}} `json:"items"`
		Total int64 `json:"total"`
	}
)


@server(
	prefix: api/{{.path}}
	group: {{.group}}
)

service app {
  // 添加
	@handler add
	post / ({{.name}})
	
	// 删除
	@handler del
	delete / (Del{{.name}}Req)

	// 获取列表
  @handler list
  get /list (List{{.name}}Req) returns (List{{.name}}Resp)

	// 更新	
	@handler update
	put / ({{.name}})
}
