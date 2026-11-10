package core

import "github.com/gofiber/fiber/v2"

func SkipAuthApi() fiber.Handler {
	// do nothing with auth
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

func SkipAuthView() fiber.Handler {
	// do nothing with auth
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}
