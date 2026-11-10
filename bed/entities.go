package bed

import "proj/imyr/config"

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
