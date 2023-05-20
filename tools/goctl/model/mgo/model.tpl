package model

import (
	"fmt"
	"context"
	"time"
	"yuqi/common/e"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

)

type {{.StructName}}Model struct {
	name      string
	primaryKey string
	secondKey string
	coll      *mongo.Collection
}

func New{{.StructName}}Model() *{{.StructName}}Model {
	name := "{{.TableName}}"
	// 创建collection
	InitCollection(name)

	m := &{{.StructName}}Model{
		name:      name,
		coll:      mdb.Collection(name),
		primaryKey: fmt.Sprintf("cache:yuqi:%s:id:", name),
		secondKey: fmt.Sprintf("cache:yuqi:%s:{{.UnisWC}}:", name),
	}
	// 创建唯一索引
	AddUniKey(m.coll, {{.UnisWDWQ}})

	return m
}

// 加载key
func (m *{{.StructName}}Model) loadKeys(data *{{.StructName}}) {
	c.Set(getKey(m.primaryKey, data.Id), data)
	c.Set(getKey(m.secondKey, {{.UnisPD}}), data.Id)
}

// 清空Key
func (m *{{.StructName}}Model) clearKeys(data *{{.StructName}}) {
	c.Del(getKey(m.primaryKey, data.Id))
	c.Del(getKey(m.secondKey, {{.UnisPD}}))
}



// 获取
func (m *{{.StructName}}Model) FindOne(id string) (data *{{.StructName}},err error) {
	data = &{{.StructName}}{}
	key := getKey(m.primaryKey, id)
	if err = c.Get(key, data); err == nil {
		return
	}

	mid, _ := primitive.ObjectIDFromHex(id)
	err = m.coll.FindOne(context.Background(), bson.M{"_id": mid}).Decode(&data)
	if data.Id != "" {
		m.loadKeys(data)
	}
	return
}

// 通过唯一key查找
func (m *{{.StructName}}Model) FindBy{{.UnisWAnd}}({{.UnisWType}}) (data *{{.StructName}}, err error) {
	data = &{{.StructName}}{}

	var id string
	key := getKey(m.secondKey, {{.UnisWD}})
	if err = c.Get(key, &id); err == nil {
		data, err = m.FindOne(id)
		return
	}

	err = m.coll.FindOne(context.Background(), bson.M{ {{.UnisWTWT}} }).Decode(data)
	if data.Id != "" {
		m.loadKeys(data)
	}
	return
}

// 创建
func (m *{{.StructName}}Model) Create(data *{{.StructName}}) (err error) {
	od, _ := m.FindBy{{.UnisWAnd}}({{.UnisPD}})
	if od.Id != "" {
		// 已经存在数据，直接返回
		err = e.AlreadyExists
		return
	}

	now := time.Now()
	data.CreateAt = now
	data.UpdateAt = now

	result, err := m.coll.InsertOne(context.Background(), data)
	if result.InsertedID != nil {
		data.Id = result.InsertedID.(primitive.ObjectID).Hex()
		m.clearKeys(data)
		return
	}
	return
}

// 更新产品
func (m *{{.StructName}}Model) Update(data *{{.StructName}}) (err error) {
	mid, _ := primitive.ObjectIDFromHex(data.Id)
	data.UpdateAt = time.Now()
	data.Id = ""
	_, err = m.coll.UpdateByID(context.Background(), mid, bson.M{"$set": data})
	if err == nil {
		m.clearKeys(data)
	}
	return
}

// 删除
func (m *{{.StructName}}Model) Delete(id string) (err error) {
	data, err := m.FindOne(id)
	if err != nil {
		logx.Infof("not found data: %+v %v", id, err)
		return
	}

	mid, _ := primitive.ObjectIDFromHex(id)
	_, err = m.coll.DeleteOne(context.Background(), bson.M{"_id": mid})
	if err == nil {
		m.clearKeys(data)
	}
	return
}

{{if .UseList}}
// 列表
func (m *{{.StructName}}Model) List(ctx context.Context, req *{{.ListStructName}}) (items []*{{.StructName}}, err error) {
	// Define the find options
	opt := options.Find()
	opt.SetSort(bson.D{{"{{"}}Key: "updateAt", Value: -1{{"}}"}})

	{{if .UsePage}}
	// 如果不是读取全部数据，就分页
	if req.Size != 0 {
	  opt.SetSkip((req.Page - 1) * req.Size)
    opt.SetLimit(req.Size)
  }
	{{end}}

	// Execute the find operation
	cursor, err := m.coll.Find(ctx, {{.ListFilter}}, opt)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	// Decode the results into a slice
	if err = cursor.All(ctx, &items); err != nil {
		return
	}

	return
}
{{end}}

func (m *{{.StructName}}Model) Count(ctx context.Context{{if .UseCountFilter}}, req *{{.ListStructName}}{{end}}) (int64, error) {
	return m.coll.CountDocuments(ctx, {{.ListFilter}})
}
