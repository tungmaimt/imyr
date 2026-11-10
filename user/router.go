package user

import (
	"github.com/gofiber/fiber/v2"
)

func Router(app fiber.Router, api fiber.Router) {
	// views
	// group := app.Group("/auth")

	// group.Get("/login", loginView())
	// group.Get("/register", registerView())

	// apis
	// apiGroup := api.Group("/auth")

	// apiGroup.Post("/login", userApi())
}

/**
 * Views
 */

/**
 * Apis
 */

// func userApi() fiber.Handler {
// 	return auth.RequireAuthenApi(func(c *fiber.Ctx) error {
// 		u := c.Locals(auth.UserKey).(auth.UserModel)

// 		userData := fiber.Map{
// 			"id":       u.ID,
// 			"email":    u.Email,
// 			"username": u.Username,
// 			"role":     u.Role,
// 			"name":     u.Username,
// 			"desc":     u.Email,
// 		}

// 		profile, err := db.FindOne[ProfileModel](map[string]any{"user": u.ID})

// 		if err == mongo.ErrNoDocuments {
// 			return c.JSON(userData)
// 		}

// 		userData["name"] = profile.Name
// 		userData["desc"] = profile.Desc

// 		return c.JSON(userData)
// 	}, &core.AuthOps{})
// }

// func loginApi() fiber.Handler {
// 	return Authen(func(c *fiber.Ctx) error {
// 		u := c.Locals("user")
// 		fmt.Println("qqqqqqqqqqqqqq")

// 		if u != nil {
// 			return c.Status(400).JSON(fiber.Map{
// 				"status":  400",
// 				"message": "already logged in",
// 			})
// 		}

// 		var d map[string]any

// 		if err := json.Unmarshal(c.Body(), &d); err != nil {
// 			return c.Status(400).JSON(fiber.Map{
// 				"status":  400,
// 				"message": "invalid json data",
// 			})
// 		}

// 		user, err := FindUsers(map[string]any{
// 			"email": d["email"],
// 		})

// 		if len(user) == 0 {
// 			return c.Status(400).JSON(fiber.Map{
// 				"status":  400,
// 				"message": "email is not registed",
// 			})
// 		}

// 		if hashPass(d["password"].(string)) != user[0].HashPass {
// 			return c.Status(400).JSON(fiber.Map{
// 				"status":  400,
// 				"message": "email or password is not corrected",
// 			})
// 		}

// 		cookie := new(fiber.Cookie)

// 		token, err := sign(jwt.MapClaims{
// 			"yid": user[0].ID.Hex(),
// 		})

// 		if err != nil {
// 			fmt.Println(err)
// 			return c.Status(500).JSON(fiber.Map{
// 				"status":  500,
// 				"message": "can not gen token",
// 			})
// 		}

// 		cookie.Name = "Ytoken"
// 		cookie.Value = token
// 		cookie.HTTPOnly = true

// 		c.Cookie(cookie)

// 		return c.JSON(fiber.Map{
// 			"status": 200,
// 			"name":   user[0].Username,
// 		})
// 	})
// }
