package user

import (
	"proj/imyr/auth"
	"proj/imyr/core"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var profileView = core.View{
	Path:              "/",
	Html:              "auth/user",
	Auth:              auth.RequireAuthenView(""),
	GetUserBindData:   auth.GetUserBindData,
	GetBindData:       func(c *fiber.Ctx) fiber.Map { return fiber.Map{"from_server": "this for render from server"} },
	GetClientBindData: func(c *fiber.Ctx) fiber.Map { return fiber.Map{"for_client": "this for access from js"} },
	ActionHander: map[string]core.ActionHandler{
		"get__profile": getProfile,
	},
}

func getProfile(c *fiber.Ctx, d map[string]any) error {
	ctx := c.Context()
	u := c.Locals(auth.UserKey).(auth.UserModel)

	userData := fiber.Map{
		"id":       u.ID,
		"email":    u.Email,
		"username": u.Username,
		"role":     u.Role,
		"name":     u.Username,
		"desc":     u.Email,
	}

	profile, err := GetProfileRepo().FindOne(ctx, map[string]any{"user": u.ID}, nil)

	if err == mongo.ErrNoDocuments {
		return c.JSON(fiber.Map{
			"status": 200,
			"data":   userData,
		})
	}

	userData["name"] = profile.Name
	userData["desc"] = profile.Desc

	return c.JSON(fiber.Map{
		"status": 200,
		"data":   userData,
	})
}
