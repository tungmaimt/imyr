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

type option func() (string, any)

const OPTION_SKIP string = "OPTION_SKIP"
const OPTION_LIMIT string = "OPTION_LIMIT"
const OPTION_PROFILE string = "OPTION_PROFILE"
const OPTION_SORT string = "OPTION_SORTT"
const OPTION_AGGREGATE string = "OPTION_AGGREGATE"

func UseSkip(offset int) option {
	return func() (string, any) {
		return OPTION_SKIP, offset
	}
}

func UseLimit(limit int) option {
	return func() (string, any) {
		return OPTION_LIMIT, limit
	}
}

func UseProfile(fields []string) option {
	return func() (string, any) {
		return OPTION_PROFILE, fields
	}
}

func UseSort(fields []string) option {
	return func() (string, any) {
		return OPTION_SORT, fields
	}
}

func UseAggregate(steps any) option {
	return func() (string, any) {
		return OPTION_AGGREGATE, steps
	}
}

func useOptions(ops ...option) any {
	queryOps := make(map[string]any)

	for _, op := range ops {
		optName, optData := op()
		switch optName {
		case OPTION_SKIP:
			queryOps["skip"] = optData
		case OPTION_LIMIT:
			queryOps["limit"] = optData
		case OPTION_PROFILE:
			queryOps["profile"] = optData
		case OPTION_SORT:
			queryOps["sort"] = optData
		case OPTION_AGGREGATE:
			queryOps["aggregate"] = optData
		}
	}

	return queryOps
}

type IndexConfig struct {
	ins    Model
	indexs []mongo.IndexModel
}

var indexCfgs = []IndexConfig{}

func AddIndex[T Model](key string, direction int, option *options.IndexOptionsBuilder) (index1 int, index2 int) {
	var ins T

	index := slices.IndexFunc(indexCfgs, func(ic IndexConfig) bool {
		return ins.GetCollectionName() == ic.ins.GetCollectionName() && ins.GetDbName() == ic.ins.GetDbName()
	})

	if index == -1 {
		indexCfgs = append(indexCfgs, IndexConfig{
			ins:    ins,
			indexs: []mongo.IndexModel{},
		})

		index = 0
	}

	indexCfgs[index].indexs = append(indexCfgs[index].indexs, mongo.IndexModel{
		Keys:    bson.D{{Key: key, Value: direction}},
		Options: option,
	})

	fmt.Printf("add index for %s: %s - %d %d\n", ins.GetCollectionName(), key, index, len(indexCfgs[index].indexs))

	return index, len(indexCfgs[index].indexs) - 1
}

func CreateIndex() (err error) {
	for _, indexConfig := range indexCfgs {
		fmt.Println("create index with ", len(indexConfig.indexs))
		getCollection(indexConfig.ins.GetCollectionName(), indexConfig.ins.GetDbName(), func(c *mongo.Collection) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			names, cerr := c.Indexes().CreateMany(ctx, indexConfig.indexs)
			err = cerr

			if err != nil {
				fmt.Print(err)
				return
			}

			fmt.Printf("Create index for %s: %v\n", indexConfig.ins.GetCollectionName(), names)
		})
	}

	return
}

func init() {
	fmt.Println("Mongo")

	client, err := mongo.Connect(
		options.Client().ApplyURI(config.DB_URL),
	)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	var result bson.M
	if err := client.Database("admin").RunCommand(
		context.TODO(),
		bson.D{{Key: "ping", Value: 1}},
	).Decode(&result); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
}

func getCollection(coll string, dbName string, fn func(*mongo.Collection)) {
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
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	database := client.Database(dbName)
	collection := database.Collection(coll)

	fn(collection)
}

func InsertOne[T Model](doc T, ops ...option) (result *mongo.InsertOneResult, err error) {
	getCollection(doc.GetCollectionName(), doc.GetDbName(), func(c *mongo.Collection) {
		result, err = c.InsertOne(context.TODO(), doc)
	})

	return
}

func InsertMany[T Model](docs []T, ops ...option) (results []mongo.InsertManyResult, err error) {
	var doc T

	getCollection(doc.GetCollectionName(), doc.GetDbName(), func(c *mongo.Collection) {
		result, cErr := c.InsertMany(context.TODO(), docs)
		err = cErr
		results = append(results, *result)
	})

	return
}

func Find[T Model](filter map[string]any, ops ...option) (docs []T, err error) {
	var docIns T

	getCollection(docIns.GetCollectionName(), docIns.GetDbName(), func(c *mongo.Collection) {
		curs, e := c.Find(context.TODO(), filter)
		if e != nil {
			err = e
			return
		}
		err = curs.All(context.TODO(), &docs)
	})

	return
}

func FindOne[T Model](filter map[string]any, ops ...option) (result *T, err error) {
	var doc T

	getCollection(doc.GetCollectionName(), doc.GetDbName(), func(c *mongo.Collection) {
		err = c.FindOne(context.TODO(), filter).Decode(&doc)
		result = &doc
	})

	return
}

func UpdateOne[T Model](filter map[string]any, update map[string]any, ops ...option) (result *mongo.UpdateResult, err error) {
	var doc T

	getCollection(doc.GetCollectionName(), doc.GetDbName(), func(c *mongo.Collection) {
		result, err = c.UpdateOne(context.TODO(), filter, update)
	})

	return
}

func UpdateMany[T Model](filter map[string]any, update map[string]any, ops ...option) (result *mongo.UpdateResult, err error) {
	var doc T

	getCollection(doc.GetCollectionName(), doc.GetDbName(), func(c *mongo.Collection) {
		result, err = c.UpdateMany(context.TODO(), filter, update)
	})

	return
}

func DeleteOne[T Model](filter map[string]any, ops ...option) (result *mongo.DeleteResult, err error) {
	var doc T

	getCollection(doc.GetCollectionName(), doc.GetDbName(), func(c *mongo.Collection) {
		result, err = c.DeleteOne(context.TODO(), filter)
	})

	return
}

func DeleteMany[T Model](filter map[string]any, ops ...option) (result *mongo.DeleteResult, err error) {
	var doc T

	getCollection(doc.GetCollectionName(), doc.GetDbName(), func(c *mongo.Collection) {
		result, err = c.DeleteMany(context.TODO(), filter)
	})

	return
}
