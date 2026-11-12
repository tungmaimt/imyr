package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"proj/imyr/config"
	"proj/imyr/core"
	"proj/imyr/db"
	"proj/imyr/services"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func hashPass(pw string) (hp string) {
	pWithN := pw + config.AUTH_HASH_PASS_NONCE
	hashBytes := sha256.Sum256([]byte(pWithN))
	hp = hex.EncodeToString(hashBytes[:])
	return
}

func checkToken(s string) (*UserModel, error) {
	var user *UserModel
	data, err := validate(s)

	if err != nil {
		return user, newAuthError(AuthErrorInvalidUser, "Invalid user")
	}

	obid, err := bson.ObjectIDFromHex(data["yid"].(string))

	if err != nil {
		return user, newAuthError(AuthErrorInvalidUser, "Invalid yid")
	}

	user, err = db.FindOne[UserModel](map[string]any{"_id": obid})

	if err == mongo.ErrNoDocuments {
		return user, newAuthError(AuthErrorNoUser, "No user")
	}

	return user, err
}

func authen(handler fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Ytoken")

		if token == "" {
			token = c.Cookies("Ytoken")
		}

		user, err := checkToken(token)

		if err == nil {
			if user.Role == ROLE_BANNED {
				c.Locals(UserKey, nil)
			} else {
				c.Locals(UserKey, user)
			}

			return handler(c)
		}

		if _, isAuthErr := err.(*AuthError); isAuthErr == true {
			return handler(c)
		}

		return fiber.NewError(500, err.Error())
	}
}

// View & Api use
// fiber.Hander = func(*fiber.Ctx) error
// ...ops = reidrect string || roles []string
// ops

// func OptionAuthenApi(handler func(*fiber.Ctx) error, ops *core.AuthOps) fiber.Handler {
// 	return authen(func(c *fiber.Ctx) error {
// 		return handler(c)
// 	}, ops)
// }

func OptionAuthenApi() fiber.Handler {
	return authen(func(c *fiber.Ctx) error {
		return c.Next()
	})
}

// func OptionAuthenView(handler fiber.Handler, ops *core.AuthOps) fiber.Handler {
// 	return authen(func(c *fiber.Ctx) error {
// 		return handler(c)
// 	}, ops)
// }

func OptionAuthenView() fiber.Handler {
	return authen(func(c *fiber.Ctx) error {
		return c.Next()
	})
}

// func NoAuthApi(handler fiber.Handler, ops *core.AuthOps) fiber.Handler {
// 	return authen(func(c *fiber.Ctx) error {
// 		if c.Locals(UserKey) == nil {
// 			return fiber.NewError(400, "Notsupport")
// 		}

// 		return handler(c)
// 	}, ops)
// }

func NoAuthApi(msg string) fiber.Handler {
	if msg == "" {
		msg = "Not supported"
	}

	return authen(func(c *fiber.Ctx) error {
		if c.Locals(UserKey) == nil {
			return c.Next()
		}

		return c.JSON(fiber.Map{
			"status":  400,
			"message": msg,
		})
	})
}

// func NoAuthView(handler fiber.Handler, ops *core.AuthOps) fiber.Handler {
// 	return authen(func(c *fiber.Ctx) error {
// 		if c.Locals(UserKey) == nil {
// 			return handler(c)
// 		}

// 		return c.Redirect(ops.Redirect)
// 	}, ops)
// }

func NoAuthView(redirect string) fiber.Handler {
	return authen(func(c *fiber.Ctx) error {
		if c.Locals(UserKey) == nil {
			return c.Next()
		}

		return c.Redirect(redirect)
	})
}

// func RequireAuthenApi(handler fiber.Handler, ops *core.AuthOps) fiber.Handler {
// 	return authen(func(c *fiber.Ctx) error {
// 		if c.Locals(UserKey) == nil {
// 			return fiber.NewError(401, "Unauthentication")
// 		}

// 		return handler(c)
// 	}, ops)
// }

func RequireAuthenApi() fiber.Handler {
	return authen(func(c *fiber.Ctx) error {
		if c.Locals(UserKey) == nil {
			return c.Status(401).JSON(fiber.Map{
				"status":  401,
				"message": "Unauthentication",
			})
		}

		return c.Next()
	})
}

// func RequireAuthenView(handler fiber.Handler, ops *core.AuthOps) fiber.Handler {
// 	return authen(func(c *fiber.Ctx) error {
// 		if c.Locals(UserKey) == nil {
// 			return c.Redirect(fmt.Sprintf("/auth/login?next=%s", url.QueryEscape(c.OriginalURL())))
// 		}

// 		return handler(c)
// 	}, ops)
// }

func RequireAuthenView(redirect string) fiber.Handler {
	return authen(func(c *fiber.Ctx) error {
		if redirect == "" {
			redirect = fmt.Sprintf("/auth/login?next=%s", url.QueryEscape(c.OriginalURL()))
		}

		if c.Locals(UserKey) == nil {
			return c.Redirect(redirect)
		}

		return c.Next()
	})
}

// func RequireAuthorApi(handler fiber.Handler, ops *core.AuthOps) fiber.Handler {
// 	return RequireAuthenApi(func(c *fiber.Ctx) error {
// 		user := c.Locals(UserKey).(UserModel)

// 		if !slices.Contains(ops.Roles, user.Role) {
// 			return fiber.NewError(403, "Unauthorization")
// 		}

// 		return handler(c)
// 	}, ops)
// }

func RequireAuthorApi(roles []string) fiber.Handler {
	return authen(func(c *fiber.Ctx) error {
		user := GetUserFromCtx(c)
		if user == nil {
			return c.Status(401).JSON(fiber.Map{
				"status":  401,
				"message": "Unauthentication",
			})
		}

		if !slices.Contains(roles, user.Role) {
			return c.Status(403).JSON(fiber.Map{
				"status":  403,
				"message": "Unauthorization",
			})
		}

		return c.Next()
	})
}

// func RequireAuthorView(handler fiber.Handler, ops *core.AuthOps) fiber.Handler {
// 	return RequireAuthenView(func(c *fiber.Ctx) error {
// 		user := c.Locals(UserKey).(UserModel)

// 		if !slices.Contains(ops.Roles, user.Role) {
// 			return c.Render("auth/403.html", core.GetBindData(fiber.Map{}))
// 		}

// 		return handler(c)
// 	}, ops)
// }

func RequireAuthorView(roles []string) fiber.Handler {
	return authen(func(c *fiber.Ctx) error {
		user := GetUserFromCtx(c)
		if user == nil {
			return c.Redirect(fmt.Sprintf("/auth/login?next=%s", url.QueryEscape(c.OriginalURL())))
		}

		if !slices.Contains(roles, user.Role) {
			return c.Render("auth/403.html", core.GetBindData(GetUserBindData(c)))
		}

		return c.Next()
	})
}

func LogoutView(redirect string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:     "Ytoken",
			Value:    "",
			Expires:  time.Now().Add(-24 * time.Hour),
			HTTPOnly: true,
			Secure:   true,
		})

		return c.Redirect(redirect)
	}
}

func LogoutApi() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:     "Ytoken",
			Value:    "",
			Expires:  time.Now().Add(-24 * time.Hour),
			HTTPOnly: true,
			Secure:   true,
		})

		return c.JSON(fiber.Map{
			"status": 200,
		})
	}
}

func GetUserFromCtx(c *fiber.Ctx) *UserModel {
	if c.Locals(UserKey) == nil {
		return nil
	}

	return c.Locals(UserKey).(*UserModel)
}

func GetUserBindData(c *fiber.Ctx) fiber.Map {
	user := GetUserFromCtx(c)

	if user == nil {
		return fiber.Map{}
	}

	return services.StructToMap(user)
}

// Db use

func FindUsers(filter map[string]any) (users []UserModel, err error) {
	users, err = db.Find[UserModel](filter)
	return
}

func insertUsers(users []UserModel) ([]any, error) {
	results, err := db.InsertMany(users)
	insertedIds := []any{}

	for _, result := range results {
		for _, insertedID := range result.InsertedIDs {
			insertedIds = append(insertedIds, insertedID)
		}
	}
	return insertedIds, err
}

func CreateUser(email string, username string, password string) (result any, err error) {
	doc := UserModel{
		Email:    email,
		Username: username,
		HashPass: hashPass(password),
		Role:     ROLE_STAFF,
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	err = services.Validator.Validate.Struct(doc)

	if err != nil {
		return nil, fmt.Errorf("validate error for: %w", err)
	}

	results, err := insertUsers([]UserModel{doc})

	return results[0], err
}

func ImportUsers() any {
	return nil
}
