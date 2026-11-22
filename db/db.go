package db

import (
	"context"
	"fmt"
	"slices"
	"time"

	"proj/imyr/config"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Model interface {
	GetCollectionName() string
	GetDbName() string
}

type MultiDbError struct{}

func (e *MultiDbError) Error() string {
	return "mongodb: multi database, should use only 1 database for each query"
}

type Repo[T Model] interface {
	InsertOne(parent context.Context, doc *T, opt *OptionBuilder) (any, error)
	FindOne(parent context.Context, filter map[string]any, opt *OptionBuilder) (*T, error)
	CreateIndex(parent context.Context, indexs []IndexModel) error
	InsertMany(parent context.Context, docs []T, opt *OptionBuilder) ([]any, error)
	Find(parent context.Context, filter map[string]any, opt *FindOptionBuilder) ([]T, error)
	UpdateOne(parent context.Context, filter map[string]any, update map[string]any, opt *OptionBuilder) (int64, error)
	UpdateMany(parent context.Context, filter map[string]any, update map[string]any, opt *OptionBuilder) (int64, error)
	DeleteOne(parent context.Context, filter map[string]any, opt *OptionBuilder) (int64, error)
	DeleteMany(parent context.Context, filter map[string]any, opt *OptionBuilder) (int64, error)
	Count(parent context.Context, filter map[string]any) (int64, error)
}

func GetRepo[T Model]() Repo[T] {
	return mongoRepo[T]{}
}

type Option struct{}

type OptionBuilder struct {
	Opts []func(opt *Option)
}

type FindOption interface {
	Skip(int)
	Limit(int)
	Projection(map[string]int) // map[name]1 map[id]0
	Sort(map[string]int)       // map[field]1 map[id]-1
}

type FindOptionBuilder struct {
	Opts []func(opt FindOption)
}

func (f *FindOptionBuilder) Skip(offset int) *FindOptionBuilder {
	f.Opts = append(f.Opts, func(opt FindOption) {
		opt.Skip(offset)
	})

	return f
}

func (f *FindOptionBuilder) Limit(limit int) *FindOptionBuilder {
	f.Opts = append(f.Opts, func(opt FindOption) {
		opt.Limit(limit)
	})

	return f
}

func (f *FindOptionBuilder) Projection(projection map[string]int) *FindOptionBuilder {
	f.Opts = append(f.Opts, func(opt FindOption) {
		opt.Projection(projection)
	})

	return f
}

func (f *FindOptionBuilder) Sort(sort map[string]int) *FindOptionBuilder {
	f.Opts = append(f.Opts, func(opt FindOption) {
		opt.Sort(sort)
	})

	return f
}

type IndexOption interface {
	Unique(bool)
	Sparse(bool)
	Text(bool)
	Geo(bool)
}

type IndexOptionsBuilder struct {
	Opts []func(IndexOption)
}

func (i *IndexOptionsBuilder) Unique(unique bool) *IndexOptionsBuilder {
	i.Opts = append(i.Opts, func(opt IndexOption) {
		opt.Unique(unique)
	})

	return i
}

func (i *IndexOptionsBuilder) Sparse(sparse bool) *IndexOptionsBuilder {
	i.Opts = append(i.Opts, func(opt IndexOption) {
		opt.Sparse(sparse)
	})

	return i
}

func (i *IndexOptionsBuilder) Text(text bool) *IndexOptionsBuilder {
	i.Opts = append(i.Opts, func(opt IndexOption) {
		opt.Text(text)
	})

	return i
}

type IndexModel struct {
	Keys   []string
	Dirs   []int
	Option *IndexOptionsBuilder
}

type IndexConfig struct {
	ins    Model
	create func(parent context.Context, indexs []IndexModel) error
	indexs []IndexModel
}

var indexCfgs = []IndexConfig{}

func AddIndex[T Model](repo Repo[T], fields []string, dirs []int, opts *IndexOptionsBuilder) (index1 int, index2 int) {
	var ins T

	index := slices.IndexFunc(indexCfgs, func(ic IndexConfig) bool {
		return ins.GetCollectionName() == ic.ins.GetCollectionName() && ins.GetDbName() == ic.ins.GetDbName()
	})

	if index == -1 {
		indexCfgs = append(indexCfgs, IndexConfig{
			ins:    ins,
			create: repo.CreateIndex,
			indexs: []IndexModel{},
		})

		index = len(indexCfgs) - 1
	}

	indexCfgs[index].indexs = append(indexCfgs[index].indexs, IndexModel{
		Keys:   fields,
		Dirs:   dirs,
		Option: opts,
	})

	fmt.Printf("add index for %s: %v - %d %d\n", ins.GetCollectionName(), fields, index, len(indexCfgs[index].indexs))

	return index, len(indexCfgs[index].indexs) - 1
}

func CreateIndex(parent context.Context) error {
	var err error
	for _, indexConfig := range indexCfgs {
		err = indexConfig.create(parent, indexConfig.indexs)
		if err != nil {
			return err
		}
	}
	return err
	// return getRepo().CreateIndex(parent, indexCfgs)
}

func init() {
	fmt.Println("Mongo")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(
		options.Client().ApplyURI(config.DB_URL),
	)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	var result bson.M
	if err := client.Database("admin").RunCommand(
		ctx,
		bson.D{{Key: "ping", Value: 1}},
	).Decode(&result); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
}
