package bed

import (
	"proj/imyr/config"
	"proj/imyr/db"
	"sync"
)

type BedModel struct {
	Name   string `json:"name"`
	Plot   string `json:"-" bson:"plot"`
	Length int    `json:"length"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Status string `json:"status"`
}

func (u BedModel) GetCollectionName() string {
	return "Bed"
}

func (BedModel) GetDbName() string {
	return config.DB_NAME
}

var repo db.Repo2[BedModel]
var once = sync.Once{}

func GetBedRepo() db.Repo2[BedModel] {
	once.Do(func() {
		repo = db.GetRepo[BedModel]()
	})

	return repo
}
