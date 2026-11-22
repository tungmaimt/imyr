package user

import (
	"proj/imyr/config"
	"proj/imyr/db"
	"sync"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// type CustomError struct {
// 	Code string
// 	Msg  string
// }

// func (e *CustomError) Error() string {
// 	return e.Msg
// }

// func newAuthError(code string, msg string) *CustomError {
// 	err := &CustomError{
// 		Code: code,
// 		Msg:  msg,
// 	}
// 	return err
// }

type ProfileModel struct {
	ID   bson.ObjectID `bson:"_id,omitempty"`
	User bson.ObjectID `bson:"user,omitempty"`
	Name string        `validate:"required,min=6,max=32"`
	Desc string        `validate:"required,max=300"`
}

func (p ProfileModel) GetCollectionName() string {
	return "Profile"
}

func (ProfileModel) GetDbName() string {
	return config.DB_NAME
}

var repo db.Repo[ProfileModel]
var once sync.Once

func GetProfileRepo() db.Repo[ProfileModel] {
	once.Do(func() {
		repo = db.GetRepo[ProfileModel]()
	})

	return repo
}
