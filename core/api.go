package core

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// LogLevel defines our custom types
type Method string

const (
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
	DELETE Method = "DELETE"
	PATCH  Method = "PATCH"
)

type Api struct {
	Path        string
	Methods     []Method
	Auth        fiber.Handler
	Handler     fiber.Handler
	Middlewares []fiber.Handler
	// Afterwares  []AfterHandler
}

func (a *Api) Register(group fiber.Router) {
	handlers := append(a.Middlewares, a.Handler)

	if a.Auth != nil {
		handlers = slices.Insert(handlers, 0, a.Auth)
	}

	for _, method := range a.Methods {

		switch method {
		case "GET":
			group.Get(a.Path, handlers...)
		case "POST":
			group.Post(a.Path, handlers...)
		case "PUT":
			group.Put(a.Path, handlers...)
		case "DELETE":
			group.Delete(a.Path, handlers...)
		case "PATCH":
			group.Patch(a.Path, handlers...)
		}
	}

}

func (a *Api) GetPage(c *fiber.Ctx) int {
	return c.QueryInt("page", 1)
}

func (a *Api) GetOffset(c *fiber.Ctx) int {
	return c.QueryInt("offset", 0)
}

func (a *Api) GetLimit(c *fiber.Ctx) int {
	return c.QueryInt("limit", 12)
}

func (a *Api) GetSize(c *fiber.Ctx) int {
	return c.QueryInt("size", 12)
}

func (a *Api) GetNext(c *fiber.Ctx, d []any) string {
	queries := c.Queries()

	_, hasOffset := queries["offset"]
	_, hasPage := queries["page"]

	if hasOffset || !hasPage {
		limit := a.GetLimit(c)
		if len(d) < limit {
			return ""
		}

		offset := c.QueryInt("offset", 0)
		queries["offset"] = fmt.Sprint(offset + limit)
	} else {
		size := a.GetSize(c)
		if len(d) < size {
			return ""
		}

		page := c.QueryInt("page", 1)
		queries["page"] = fmt.Sprint(page + 1)
	}

	querySlice := []string{}

	for k, v := range queries {
		querySlice = append(querySlice, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
	}

	return fmt.Sprintf("%s?%s", c.Path, strings.Join(querySlice, "&"))
}
