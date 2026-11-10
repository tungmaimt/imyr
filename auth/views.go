package auth

import (
	"proj/imyr/core"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

/**
 * Login
 */

var loginView = core.View{
	Path: "/login",
	Html: "auth/login",
	Auth: NoAuthView("/"),
	ActionHander: map[string]core.ActionHandler{
		"login": actionLogin,
	},
	Middlewares: []fiber.Handler{core.LogURL, core.LogReq},
}

func actionLogin(c *fiber.Ctx, d map[string]any) error {
	username := d["username"]
	password := d["password"]

	user, err := FindUsers(map[string]any{"username": username})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  500,
			"message": "system error, try another time",
		})
	}

	if len(user) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"status":  400,
			"message": "username is not regist",
		})
	}

	if hashPass(password.(string)) != user[0].HashPass {
		return c.Status(400).JSON(fiber.Map{
			"status":  400,
			"message": "usnername or password is not correct",
		})
	}

	token, err := sign(jwt.MapClaims{
		"yid": user[0].ID.Hex(),
	})

	cookie := new(fiber.Cookie)

	cookie.Name = "Ytoken"
	cookie.Value = token
	cookie.HTTPOnly = true
	cookie.Secure = true

	c.Cookie(cookie)

	return c.JSON(fiber.Map{
		"status": 200,
		"name":   user[0].Username,
	})
}

/**
 * Register
 */

// var registerView = core.View{
// 	Path: "register",
// 	Html: "auth/register",
// 	Auth: NoAuthView(),
// 	ActionHander: map[string]core.ActionHandler{
// 		"register": actionRegister,
// 	},
// }

/**
 * Logout
 */

var logoutView = core.View{
	Path: "/logout",
	Html: "auth/logout",
	Auth: LogoutView("/"),
}
