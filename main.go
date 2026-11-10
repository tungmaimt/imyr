package main

import (
	"proj/imyr/auth"
	"proj/imyr/bed"
	"proj/imyr/config"
	"proj/imyr/core"
	"proj/imyr/user"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/django/v3"
)

func main() {
	/**
	 * django/pongo2 view engine
	 */

	view := django.New("./views", ".html")

	if config.ENV == config.ENV_LOCAL {
		view.Reload(true)
	}

	/**
	 * fiber app
	 */

	app := fiber.New(fiber.Config{
		EnableSplittingOnParsers: true,
		Views:                    view,
	})

	staticConfig := fiber.Static{}

	if config.ENV == config.ENV_LOCAL {
		staticConfig.CacheDuration = -1
		staticConfig.MaxAge = -1
	}

	app.Static(config.STATIC_PATH, config.STATIC_DIR, staticConfig)
	app.Static(config.PUBLIC_PATH, config.PUBLIC_DIR, staticConfig)

	api := app.Group("/api")

	// regist sub apps router
	auth.Router(app, api)
	user.Router(app, api)

	// test
	bed.Router(app, api)

	// homepage
	homeView := core.View{
		Path:            "/",
		Html:            "main",
		Auth:            auth.OptionAuthenView(),
		GetUserBindData: auth.GetUserBindData,
	}

	homeView.Register(app)

	// app.Post("/api/test", func(c *fiber.Ctx) error {
	// 	var d map[string]any

	// 	if err := json.Unmarshal(c.Body(), &d); err != nil {
	// 		fmt.Println("error", err)
	// 	}
	// 	fmt.Printf("%T %v \n", d, d["some"] == "thing")
	// 	fmt.Println("json data", d["some"])

	// 	nest := d["nest"].(map[string]any)

	// 	arr := []int{}

	// 	for _, v := range nest["arr"].([]any) {
	// 		arr = append(arr, int(v.(float64)))
	// 	}

	// 	println("nest", nest["v"], int(nest["arr"].([]any)[0].(float64)))
	// 	println("arr", arr[0], arr[1] == 2)

	// 	c.JSON(fiber.Map{
	// 		"data": "ok",
	// 	})

	// 	return nil
	// })

	// app.All("/error", func(c *fiber.Ctx) error {
	// 	// return c.Status(400).JSON(fiber.Map{
	// 	// 	"status": "error",
	// 	// 	"message": "This is error",
	// 	// })
	// 	return fiber.NewError(400, "This is error")
	// })

	/**
	 * mongo client
	 */

	// fmt.Println("Mongo")

	// client, err := mongo.Connect(
	// 	options.Client().ApplyURI(
	// 		"mongodb://tungmai:tungmai@localhost:27017",
	// 	),
	// )

	// if err != nil {
	// 	panic(err)
	// }

	// defer func() {
	// 	if err := client.Disconnect(context.TODO()); err != nil {
	// 		panic(err)
	// 	}
	// }()

	// var result bson.M
	// if err := client.Database("admin").RunCommand(
	// 	context.TODO(),
	// 	bson.D{{Key: "ping", Value: 1}},
	// ).Decode(&result); err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	/**
	 * start
	 */

	err := app.Listen(":3000")

	if err != nil {
		panic(err)
	}
}
