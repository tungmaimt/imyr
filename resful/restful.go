package restful

import (
	"encoding/json"
	"fmt"
	"proj/imyr/config"
	"proj/imyr/core"
	"proj/imyr/db"
	"proj/imyr/services"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Filter struct {
	Filters []string `query:"filters"`
}

func RestFul[T db.Model](group fiber.Router, methods []core.Method, auth fiber.Handler, middlewares []fiber.Handler, repo db.Repo2[T]) {
	var ins T
	path := strings.ToLower(ins.GetCollectionName())
	idPath := fmt.Sprintf("%s/:id", path)
	apis := make([]core.Api, 6)

	for _, method := range methods {
		switch method {
		case core.GET:
			apis = append(
				apis,
				core.Api{
					Path:        idPath,
					Methods:     []core.Method{method},
					Auth:        auth,
					Middlewares: middlewares,
					Handler: func(c *fiber.Ctx) error {
						ctx := c.Context()
						objId, _ := bson.ObjectIDFromHex(c.Params("id"))

						doc, err := repo.FindOne(
							ctx, map[string]any{
								"_id": objId,
							},
							nil,
						)

						if err == mongo.ErrNoDocuments {
							return c.JSON(fiber.Map{
								"status": 200,
								"data":   nil,
							})
						}

						if err != nil {
							return c.JSON(fiber.Map{
								"status": 500,
								"data":   nil,
							})
						}

						return c.JSON(fiber.Map{
							"status": 200,
							"data":   services.StructToMap(doc),
						})
					},
				},
				core.Api{
					Path:        path,
					Methods:     []core.Method{method},
					Auth:        auth,
					Middlewares: middlewares,
					Handler: func(c *fiber.Ctx) error {
						ctx := c.Context()
						filter := make(map[string]any)
						filterStr := new(Filter)
						err := c.QueryParser(filterStr)

						if err != nil {
							fmt.Print("resful get many query parse error", err)
							// TODO: maybe res with 400?
						} else {
							filter = config.FilterFromQueryString(filterStr.Filters)
						}

						docs, err := repo.Find(ctx, filter, nil)

						if err != nil {
							fmt.Println("restfull RestFUL db.Find err", err)
							return c.Status(500).JSON(fiber.Map{
								"status":  500,
								"message": "something went wrong",
							})
						}

						data := []map[string]any{}

						for _, doc := range docs {
							data = append(data, services.StructToMap(doc))
						}

						return c.JSON(fiber.Map{
							"status": 200,
							"data":   data,
						})
					},
				},
			)
		case core.POST:
			apis = append(apis, core.Api{
				Path:        path,
				Methods:     []core.Method{method},
				Auth:        auth,
				Middlewares: middlewares,
				Handler: func(c *fiber.Ctx) error {
					ctx := c.Context()
					var d map[string]any

					if err := json.Unmarshal(c.Body(), &d); err != nil {
						return c.Status(400).JSON(fiber.Map{
							"status":  400,
							"message": "invalid json data",
						})
					}

					d["status"] = config.STATUS_ACTIVE

					_, haveStatus := d["status"]

					if !haveStatus {
						d["status"] = config.STATUS_ACTIVE
					}

					ins, err := services.MapToStruct[T](d)

					if err != nil {
						return c.Status(400).JSON(fiber.Map{
							"status":  400,
							"message": err.Error(),
						})
					}

					result, err := repo.InsertOne(ctx, ins, nil)

					if err != nil {
						return c.Status(400).JSON(fiber.Map{
							"status":  400,
							"message": err.Error(),
						})
					}

					if result == nil {
						fmt.Println(err, result, "Error create with restful", c.OriginalURL())
						return c.Status(500).JSON(fiber.Map{
							"status":  500,
							"message": fmt.Sprintf("Create unsuccessful for api %s", path),
						})
					}

					return c.JSON(fiber.Map{
						"status": 200,
						"data":   result,
					})
				},
			})
		case core.PUT:
			apis = append(apis, core.Api{
				Path:        idPath,
				Methods:     []core.Method{method},
				Auth:        auth,
				Middlewares: middlewares,
				Handler: func(c *fiber.Ctx) error {
					ctx := c.Context()
					var d map[string]any

					objId, err := bson.ObjectIDFromHex(c.Params("id"))

					if err := json.Unmarshal(c.Body(), &d); err != nil {
						return c.Status(400).JSON(fiber.Map{
							"status":  400,
							"message": "invalid json data",
						})
					}

					delete(d, "status")

					fmt.Print(c.Params("id"), "--restful--", d)

					result, err := repo.UpdateOne(ctx, map[string]any{"_id": objId}, map[string]any{"$set": d}, nil)

					if err != nil {
						return c.Status(400).JSON(fiber.Map{
							"status":  400,
							"message": err.Error(),
						})
					}

					return c.JSON(fiber.Map{
						"status": 200,
						"data":   result,
					})
				},
			})
		case core.DELETE:
			apis = append(apis, core.Api{
				Path:        idPath,
				Methods:     []core.Method{method},
				Auth:        auth,
				Middlewares: middlewares,
				Handler: func(c *fiber.Ctx) error {
					ctx := c.Context()
					objId, err := bson.ObjectIDFromHex(c.Params("id"))

					result, err := repo.UpdateOne(
						ctx,
						map[string]any{"_id": objId},
						map[string]any{"$set": map[string]any{"status": config.STATUS_DELETED}},
						nil,
					)

					if err != nil {
						return c.Status(400).JSON(fiber.Map{
							"status":  400,
							"message": err.Error(),
						})
					}

					return c.JSON(fiber.Map{
						"status": 200,
						"data":   result,
					})
				},
			})
		}
	}

	for _, api := range apis {
		api.Register(group)
	}
}
