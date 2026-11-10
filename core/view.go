package core

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"proj/imyr/config"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ActionHandler func(c *fiber.Ctx, d map[string]any) error

var defaultBindData = fiber.Map{
	"static_template_url": config.STATIC_TEMPLATE_PATH,
	"static_url":          config.STATIC_PATH,
	"static_version":      time.Now().UnixMilli(),
}

func GetBindData(extData fiber.Map) fiber.Map {
	maps.Copy(extData, defaultBindData)
	return extData
}

func getEmptyBinData(c *fiber.Ctx) fiber.Map {
	return fiber.Map{}
}

type View struct {
	Path              string
	Html              string
	GetBindData       func(c *fiber.Ctx) fiber.Map
	GetUserBindData   func(c *fiber.Ctx) fiber.Map
	GetClientBindData func(c *fiber.Ctx) fiber.Map
	Auth              fiber.Handler
	ActionHander      map[string]ActionHandler
	Middlewares       []fiber.Handler
	// Afterwares        []AfterHandler
}

func (v *View) Register(group fiber.Router) {
	// if v.Auth == nil {
	// 	v.Auth = SkipAuthView()
	// }

	if v.GetBindData == nil {
		v.GetBindData = getEmptyBinData
	}

	if v.GetUserBindData == nil {
		v.GetUserBindData = getEmptyBinData
	}

	if v.GetClientBindData == nil {
		v.GetClientBindData = getEmptyBinData
	}

	render := func(c *fiber.Ctx) error {
		clientData := v.GetClientBindData(c)
		clientJson, err := json.Marshal(clientData)
		clientStr := fmt.Sprintf("{\"error\": \"fail convert json client data - %s\"}", err)

		if err == nil {
			clientStr = string(clientJson)
		}

		userData := v.GetUserBindData(c)
		userJson, err := json.Marshal(userData)
		userStr := fmt.Sprintf("{\"error\": \"fail convert json user data - %s\"}", err)

		if err == nil {
			userStr = string(userJson)
		}

		bindData := GetBindData(v.GetBindData(c))

		// add user bind data and client bind data
		maps.Copy(bindData, fiber.Map{
			"user":        userData,
			"yuser":       userStr,
			"client_data": clientStr,
		})

		c.Render(v.Html, bindData)
		return nil
	}

	action := func(c *fiber.Ctx) error {
		var d map[string]any

		if err := json.Unmarshal(c.Body(), &d); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  400,
				"message": "invalid json data",
			})
		}

		actionInfo, _ := d["action"]

		if actionInfo == nil {
			return c.Status(400).JSON(fiber.Map{
				"stauts":  400,
				"message": "action is required",
			})
		}

		actionKey, ok := actionInfo.(string)

		if !ok {
			return c.Status(400).JSON(fiber.Map{
				"stauts":  400,
				"message": "invalid action",
			})
		}

		handler, _ := v.ActionHander[actionKey]

		if handler == nil {
			return c.Status(400).JSON(fiber.Map{
				"stauts":  400,
				"message": "invalid action",
			})
		}

		return handler(c, d)
	}

	getHandlers := append(v.Middlewares, render)
	postHandlers := append(v.Middlewares, action)

	if v.Auth != nil {
		getHandlers = slices.Insert(getHandlers, 0, v.Auth)
		postHandlers = slices.Insert(postHandlers, 0, v.Auth)
	}

	group.Get(v.Path, getHandlers...)
	group.Post(v.Path, postHandlers...)
}

func (v *View) GetPage(c *fiber.Ctx, d map[string]any) int {
	return d["page"].(int)
}

func (v *View) GetOffset(c *fiber.Ctx, d map[string]any) int {
	return d["offset"].(int)
}

func (v *View) GetLimit(c *fiber.Ctx, d map[string]any) int {
	return d["limit"].(int)
}

func (v *View) GetSize(c *fiber.Ctx, d map[string]any) int {
	return d["size"].(int)
}

func (v *View) GetNext(c *fiber.Ctx, d map[string]any, data []any) string {
	queries := c.Queries()

	_, hasOffset := queries["offset"]
	_, hasPage := queries["page"]

	if hasOffset || !hasPage {
		limit := v.GetLimit(c, d)
		if len(data) < limit {
			return ""
		}

		offset := c.QueryInt("offset", 0)
		queries["offset"] = fmt.Sprint(offset + limit)
	} else {
		size := v.GetSize(c, d)
		if len(data) < size {
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
