package db

import (
	"context"
	"fmt"
	"proj/imyr/config"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func getCollection(parent context.Context, coll string, dbName string, fn func(context.Context, *mongo.Collection)) {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()

	if coll == "" {
		panic("mongodb: no coll")
	}

	if dbName == "" {
		panic("mongodb: no database")
	}

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

	database := client.Database(dbName)
	collection := database.Collection(coll)

	fn(ctx, collection)
}

type mongoIndexOption struct {
	Builder   *options.IndexOptionsBuilder
	TextIndex bool
	GeoIndex  bool
}

func (m *mongoIndexOption) Unique(unique bool) {
	m.Builder = m.Builder.SetUnique(unique)
}

func (m *mongoIndexOption) Sparse(sparse bool) {
	m.Builder = m.Builder.SetSparse(sparse)
}

func (m *mongoIndexOption) Text(text bool) {
	m.TextIndex = text
}

func (m *mongoIndexOption) Geo(geo bool) {
	m.GeoIndex = geo
}

func convertIndexs(indexs []IndexModel) []mongo.IndexModel {
	data := []mongo.IndexModel{}

	for _, index := range indexs {
		keys := bson.D{}

		mongoIndexOpt := &mongoIndexOption{Builder: &options.IndexOptionsBuilder{}}

		if index.Option != nil {
			for _, optFunc := range index.Option.Opts {
				optFunc(mongoIndexOpt)
			}
		}

		for i, k := range index.Keys {
			var val any = index.Dirs[i]
			if mongoIndexOpt.TextIndex {
				val = "text"
			} else if mongoIndexOpt.GeoIndex {
				val = "2d"
			} else if val != 1 && val != -1 {
				val = -1
			}

			keys = append(keys, bson.E{Key: k, Value: val})
		}

		data = append(data, mongo.IndexModel{
			Keys:    keys,
			Options: mongoIndexOpt.Builder,
		})
	}

	return data
}

type mongoFindOption struct {
	Builder *options.FindOptionsBuilder
}

func (m *mongoFindOption) Skip(offset int) {
	m.Builder = m.Builder.SetSkip(int64(offset))
}

func (m *mongoFindOption) Limit(limit int) {
	m.Builder = m.Builder.SetLimit(int64(limit))
}

func (m *mongoFindOption) Projection(projection map[string]int) {
	data := bson.D{}
	for k, v := range projection {
		data = append(data, bson.E{
			Key:   k,
			Value: v,
		})
	}

	m.Builder = m.Builder.SetProjection(data)
}

func (m *mongoFindOption) Sort(sort map[string]int) {
	data := bson.D{}
	for k, v := range sort {
		data = append(data, bson.E{
			Key:   k,
			Value: v,
		})
	}

	m.Builder = m.Builder.SetSort(data)
}

func convertFindOptions(opt *FindOptionBuilder) *options.FindOptionsBuilder {
	mongoFindOpt := &mongoFindOption{Builder: &options.FindOptionsBuilder{}}

	if opt != nil {
		for _, optFunc := range opt.Opts {
			optFunc(mongoFindOpt)
		}
	}

	return mongoFindOpt.Builder
}

type mongoRepo[T Model] struct {
}

func (mongoRepo[T]) CreateIndex(parent context.Context, indexs []IndexModel, ins Model) (err error) {
	fmt.Println("create index with ", ins.GetCollectionName(), len(indexs))
	getCollection(
		parent,
		ins.GetCollectionName(),
		ins.GetDbName(),
		func(ctx context.Context, c *mongo.Collection) {
			names, cerr := c.Indexes().CreateMany(ctx, convertIndexs(indexs))
			err = cerr

			if err != nil {
				fmt.Print(err)
				return
			}

			fmt.Printf("Create index for %s: %v\n", ins.GetCollectionName(), names)
		},
	)

	return
}

func (mongoRepo[T]) InsertOne(parent context.Context, coll string, dbname string, doc T, opt *OptionBuilder) (any, error) {
	var result *mongo.InsertOneResult
	var err error

	getCollection(parent, coll, dbname, func(ctx context.Context, c *mongo.Collection) {
		result, err = c.InsertOne(ctx, doc)
	})

	return result.InsertedID, err
}

func (mongoRepo[T]) InsertMany(parent context.Context, coll string, dbname string, docs []T, opt *OptionBuilder) ([]any, error) {
	var result *mongo.InsertManyResult
	var err error

	getCollection(parent, coll, dbname, func(ctx context.Context, c *mongo.Collection) {
		result, err = c.InsertMany(ctx, docs)
	})

	return result.InsertedIDs, err
}

func (mongoRepo[T]) Find(parent context.Context, coll string, dbname string, docs []T, filter map[string]any, opt *FindOptionBuilder) ([]T, error) {
	var err error
	var val []T

	getCollection(parent, coll, dbname, func(ctx context.Context, c *mongo.Collection) {
		curs, err := c.Find(ctx, filter, convertFindOptions(opt))
		if err != nil {
			return
		}
		err = curs.All(ctx, &val)
	})

	return val, err
}

func (mongoRepo[T]) FindOne(parent context.Context, coll string, dbname string, doc *T, filter map[string]any, opt *OptionBuilder) (*T, error) {
	var err error
	var val T

	getCollection(parent, coll, dbname, func(ctx context.Context, c *mongo.Collection) {
		err = c.FindOne(ctx, filter).Decode(&val)
	})

	return &val, err
}

func (mongoRepo[T]) UpdateOne(parent context.Context, coll string, dbname string, filter map[string]any, update map[string]any, opt *OptionBuilder) (int64, error) {
	var result *mongo.UpdateResult
	var err error

	getCollection(parent, coll, dbname, func(ctx context.Context, c *mongo.Collection) {
		result, err = c.UpdateOne(ctx, filter, update)
	})

	return result.ModifiedCount, err
}

func (mongoRepo[T]) UpdateMany(parent context.Context, coll string, dbname string, filter map[string]any, update map[string]any, opt *OptionBuilder) (int64, error) {
	var result *mongo.UpdateResult
	var err error

	getCollection(parent, coll, dbname, func(ctx context.Context, c *mongo.Collection) {
		result, err = c.UpdateMany(ctx, filter, update)
	})

	return result.ModifiedCount, err
}

func (mongoRepo[T]) DeleteOne(parent context.Context, coll string, dbname string, filter map[string]any, opt *OptionBuilder) (int64, error) {
	var result *mongo.DeleteResult
	var err error

	getCollection(parent, coll, dbname, func(ctx context.Context, c *mongo.Collection) {
		result, err = c.DeleteOne(ctx, filter)
	})

	return result.DeletedCount, err
}

func (mongoRepo[T]) DeleteMany(parent context.Context, coll string, dbname string, filter map[string]any, opt *OptionBuilder) (int64, error) {
	var result *mongo.DeleteResult
	var err error

	getCollection(parent, coll, dbname, func(ctx context.Context, c *mongo.Collection) {
		result, err = c.DeleteMany(ctx, filter)
	})

	return result.DeletedCount, err
}

func (mongoRepo[T]) Count(parent context.Context, coll string, dbname string, filter map[string]any) (int64, error) {
	var result int64
	var err error
	getCollection(parent, coll, dbname, func(ctx context.Context, c *mongo.Collection) {
		result, err = c.CountDocuments(ctx, filter)
	})

	return result, err
}
