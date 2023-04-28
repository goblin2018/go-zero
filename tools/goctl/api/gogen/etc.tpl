Name: {{.serviceName}}
Host: {{.host}}
Port: {{.port}}

Log:
  ServiceName: {{.serviceName}}-api
  Encoding: plain
  TimeFormat: 15:04:05.000

Redis:
  Host: redis:6379
  Type: node

DB:
  DataSource: root:xy2089@tcp(mysql:3306)/yuqi?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai

Mongo:
  Url: mongodb://root:xy2089@mongo:27017

Cache:
  - Host: redis:6379
  
Auth:
  Key: bezDoiKl3ffamiPqVNTML28VwjY
  Expire: 15

