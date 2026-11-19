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

type Repo[T Model] interface {
	CreateIndex(parent context.Context, indexs []IndexModel, ins Model) error
	InsertOne(parent context.Context, coll string, dbname string, doc T, opt *OptionBuilder) (any, error)
	InsertMany(parent context.Context, coll string, dbname string, docs []T, opt *OptionBuilder) ([]any, error)
	Find(parent context.Context, coll string, dbname string, docs []T, filter map[string]any, opt *FindOptionBuilder) ([]T, error)
	FindOne(parent context.Context, coll string, dbname string, doc *T, filter map[string]any, opt *OptionBuilder) (*T, error)
	UpdateOne(parent context.Context, coll string, dbname string, filter map[string]any, update map[string]any, opt *OptionBuilder) (int64, error)
	UpdateMany(parent context.Context, coll string, dbname string, filter map[string]any, update map[string]any, opt *OptionBuilder) (int64, error)
	DeleteOne(parent context.Context, coll string, dbname string, filter map[string]any, opt *OptionBuilder) (int64, error)
	DeleteMany(parent context.Context, coll string, dbname string, filter map[string]any, opt *OptionBuilder) (int64, error)
	Count(parent context.Context, coll string, dbname string, filter map[string]any) (int64, error)
}

func getRepo[T Model]() Repo[T] {
	repo := mongoRepo[T]{}
	return repo
}

type MultiDbError struct{}

func (e *MultiDbError) Error() string {
	return "mongodb: multi database, should use only 1 database for each query"
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
	create func(parent context.Context, indexs []IndexModel, ins Model) error
	indexs []IndexModel
}

var indexCfgs = []IndexConfig{}

func AddIndex[T Model](fields []string, dirs []int, opts *IndexOptionsBuilder) (index1 int, index2 int) {
	var ins T

	index := slices.IndexFunc(indexCfgs, func(ic IndexConfig) bool {
		return ins.GetCollectionName() == ic.ins.GetCollectionName() && ins.GetDbName() == ic.ins.GetDbName()
	})

	if index == -1 {
		indexCfgs = append(indexCfgs, IndexConfig{
			ins:    ins,
			create: getRepo[T]().CreateIndex,
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
		err = indexConfig.create(parent, indexConfig.indexs, indexConfig.ins)
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

func InsertOne[T Model](parent context.Context, doc T, opt *OptionBuilder) (any, error) {
	return getRepo[T]().InsertOne(parent, doc.GetCollectionName(), doc.GetDbName(), doc, opt)
}

// func InsertOne[T Model](parent context.Context, doc T, opt ...option) (result *mongo.InsertOneResult, err error) {
// 	getCollection(parent, doc.GetCollectionName(), doc.GetDbName(), func(ctx context.Context, c *mongo.Collection) {
// 		result, err = c.InsertOne(ctx, doc)
// 	})

// 	return
// }

func InsertMany[T Model](parent context.Context, docs []T, opt *OptionBuilder) ([]any, error) {
	var doc T

	mDocs := make([]T, len(docs))
	copy(mDocs, docs)

	return getRepo[T]().InsertMany(parent, doc.GetCollectionName(), doc.GetDbName(), mDocs, opt)
}

// func InsertMany[T Model](parent context.Context, docs []T, opt ...option) (results []mongo.InsertManyResult, err error) {
// 	var doc T

// 	getCollection(parent, doc.GetCollectionName(), doc.GetDbName(), func(ctx context.Context, c *mongo.Collection) {
// 		result, cErr := c.InsertMany(ctx, docs)
// 		err = cErr
// 		results = append(results, *result)
// 	})

// 	return
// }

func Find[T Model](parent context.Context, filter map[string]any, opt *FindOptionBuilder) ([]T, error) {
	var doc T
	var docs []T

	docs, err := getRepo[T]().Find(parent, doc.GetCollectionName(), doc.GetDbName(), docs, filter, opt)

	// result := make([]T, len(docs))
	// for i, d := range docs {
	// 	val, ok := d.(T)
	// 	if !ok {
	// 		return result, fmt.Errorf("elem at index %d as incorrect type %T", i, d)
	// 	}

	// 	result[i] = val
	// }

	return docs, err
}

// func Find[T Model](parent context.Context, filter map[string]any, opt ...FindOption) (docs []T, err error) {
// 	var docIns T

// 	getCollection(parent, docIns.GetCollectionName(), docIns.GetDbName(), func(ctx context.Context, c *mongo.Collection) {
// 		curs, e := c.Find(ctx, filter, useFindOptions(opt...))
// 		if e != nil {
// 			err = e
// 			return
// 		}
// 		err = curs.All(ctx, &docs)
// 	})

// 	return
// }

func FindOne[T Model](parent context.Context, filter map[string]any, opt *OptionBuilder) (*T, error) {
	var doc T

	result, err := getRepo[T]().FindOne(parent, doc.GetCollectionName(), doc.GetDbName(), &doc, filter, opt)

	// fmt.Println("rrrrrrrrrrrrr", result)
	// fmt.Printf("%T\n", result)

	// doc, ok := result.(T)
	// if !ok {
	// 	return nil, fmt.Errorf("incorrect type %T", result)
	// }

	return result, err
}

// func FindOne[T Model](parent context.Context, filter map[string]any, opt ...option) (result *T, err error) {
// 	var doc T

// 	getCollection(parent, doc.GetCollectionName(), doc.GetDbName(), func(ctx context.Context, c *mongo.Collection) {
// 		err = c.FindOne(ctx, filter).Decode(&doc)
// 		result = &doc
// 	})

// 	return
// }

func UpdateOne[T Model](parent context.Context, filter map[string]any, update map[string]any, opt *OptionBuilder) (int64, error) {
	var doc T
	return getRepo[T]().UpdateOne(parent, doc.GetCollectionName(), doc.GetDbName(), filter, update, opt)
}

// func UpdateOne[T Model](parent context.Context, filter map[string]any, update map[string]any, opt ...option) (result *mongo.UpdateResult, err error) {
// 	var doc T

// 	getCollection(parent, doc.GetCollectionName(), doc.GetDbName(), func(ctx context.Context, c *mongo.Collection) {
// 		result, err = c.UpdateOne(ctx, filter, update)
// 	})

// 	return
// }

func UpdateMany[T Model](parent context.Context, filter map[string]any, update map[string]any, opt *OptionBuilder) (int64, error) {
	var doc T
	return getRepo[T]().UpdateMany(parent, doc.GetCollectionName(), doc.GetDbName(), filter, update, opt)
}

// func UpdateMany[T Model](parent context.Context, filter map[string]any, update map[string]any, opt ...option) (result *mongo.UpdateResult, err error) {
// 	var doc T

// 	getCollection(parent, doc.GetCollectionName(), doc.GetDbName(), func(ctx context.Context, c *mongo.Collection) {
// 		result, err = c.UpdateMany(ctx, filter, update)
// 	})

// 	return
// }

func DeleteOne[T Model](parent context.Context, filter map[string]any, opt *OptionBuilder) (int64, error) {
	var doc T
	return getRepo[T]().DeleteOne(parent, doc.GetCollectionName(), doc.GetDbName(), filter, opt)
}

// func DeleteOne[T Model](parent context.Context, filter map[string]any, opt ...option) (result *mongo.DeleteResult, err error) {
// 	var doc T

// 	getCollection(parent, doc.GetCollectionName(), doc.GetDbName(), func(ctx context.Context, c *mongo.Collection) {
// 		result, err = c.DeleteOne(ctx, filter)
// 	})

// 	return
// }

func DeleteMany[T Model](parent context.Context, filter map[string]any, opt *OptionBuilder) (int64, error) {
	var doc T
	return getRepo[T]().DeleteMany(parent, doc.GetCollectionName(), doc.GetDbName(), filter, opt)
}

// func DeleteMany[T Model](parent context.Context, filter map[string]any, opt ...option) (result *mongo.DeleteResult, err error) {
// 	var doc T

// 	getCollection(parent, doc.GetCollectionName(), doc.GetDbName(), func(ctx context.Context, c *mongo.Collection) {
// 		result, err = c.DeleteMany(ctx, filter)
// 	})

// 	return
// }

func Count[T Model](parent context.Context, filter map[string]any) (int64, error) {
	var doc T
	return getRepo[T]().Count(parent, doc.GetCollectionName(), doc.GetDbName(), filter)
}

// func Count[T Model](parent context.Context, filter map[string]any) (result int64, err error) {
// 	var doc T
// 	getCollection(parent, doc.GetCollectionName(), doc.GetDbName(), func(ctx context.Context, c *mongo.Collection) {
// 		result, err = c.CountDocuments(ctx, filter)
// 	})

// 	return
// }
