package bed

import (
	"proj/imyr/auth"
	"proj/imyr/core"
	restful "proj/imyr/resful"

	"github.com/gofiber/fiber/v2"
)

func Router(app fiber.Router, api fiber.Router) {
	restful.RestFul[BedModel](
		api,
		[]core.Method{core.GET, core.POST, core.PUT, core.DELETE},
		auth.RequireAuthenApi(),
		[]fiber.Handler{core.LogURL, core.LogReq},
	)
}
