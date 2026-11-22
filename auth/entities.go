package auth

import (
	"sync"
	"time"

	"proj/imyr/config"
	"proj/imyr/db"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const AuthErrorInvalidUser = "invalid_user"
const AuthErrorNoUser = "no_user"

type LocalKey string

const UserKey LocalKey = "auth_local_key_user"

type AuthError struct {
	Code string
	Msg  string
}

func (e *AuthError) Error() string {
	return e.Msg
}

func newAuthError(code string, msg string) *AuthError {
	err := &AuthError{
		Code: code,
		Msg:  msg,
	}
	return err
}

type UserModel struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"_id"`
	Email    string        `validate:"omitempty,email"`
	Username string        `validate:"required,min=6,max=32"`
	HashPass string        `bson:"hash_pass" json:"-"`
	Role     string
	Created  time.Time `json:"-"`
	Updated  time.Time `json:"-"`
}

func (u UserModel) GetCollectionName() string {
	return "User"
}

func (UserModel) GetDbName() string {
	return config.DB_NAME
}

var repo db.Repo[UserModel]
var once = sync.Once{}

func GetUserRepo() db.Repo[UserModel] {
	once.Do(func() {
		repo = db.GetRepo[UserModel]()
	})

	return repo
}

const ROLE_ADMIN = "admin"
const ROLE_STAFF = "staff"
const ROLE_BANNED = "banned"
