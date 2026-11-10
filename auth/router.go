package auth

import (
	"github.com/gofiber/fiber/v2"
)

func Router(app fiber.Router, api fiber.Router) {
	// views
	group := app.Group("/auth")

	loginView.Register(group)
	logoutView.Register(group)

	// apis
	apiGroup := api.Group("/auth")

	loginApi.Register(apiGroup)
	logoutApi.Register(apiGroup)

}
