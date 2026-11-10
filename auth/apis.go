package auth

import (
	"encoding/json"
	"fmt"
	"proj/imyr/core"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

/**
 * Apis
 */

var loginApi = core.Api{
	Path:        "/login",
	Methods:     []core.Method{core.POST},
	Auth:        NoAuthApi("already loged in"),
	Middlewares: []fiber.Handler{core.LogURL, core.LogReq},
	Handler: func(c *fiber.Ctx) error {
		var d map[string]any

		if err := json.Unmarshal(c.Body(), &d); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  400,
				"message": "invalid json data",
			})
		}

		user, err := FindUsers(map[string]any{
			"email": d["email"],
		})

		if err != nil {
			fmt.Println(err)
			return c.Status(500).JSON(fiber.Map{
				"status":  500,
				"message": "system error, try another time",
			})
		}

		if len(user) == 0 {
			return c.Status(400).JSON(fiber.Map{
				"status":  400,
				"message": "email is not registed",
			})
		}

		if hashPass(d["password"].(string)) != user[0].HashPass {
			return c.Status(400).JSON(fiber.Map{
				"status":  400,
				"message": "email or password is not corrected",
			})
		}

		token, err := sign(jwt.MapClaims{
			"yid": user[0].ID.Hex(),
		})

		if err != nil {
			fmt.Printf("error login api", err)
			return c.JSON(fiber.Map{
				"status": 500,
				"name":   "Something went wrong",
			})
		}

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
	},
}

var logoutApi = core.Api{
	Path:        "/logout",
	Methods:     []core.Method{core.GET, core.POST, core.PUT, core.DELETE, core.PATCH},
	Middlewares: []fiber.Handler{core.LogURL, core.LogReq},
	Handler:     logoutHandler,
}

func logoutHandler(c *fiber.Ctx) error {
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
