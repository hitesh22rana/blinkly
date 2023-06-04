package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hitesh22rana/blinkly/database"
	"github.com/redis/go-redis/v9"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url")

	rdb := database.CreateClient(0)
	defer rdb.Close()

	val, err := rdb.Get(database.Ctx, url).Result()

	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "URL does not exist",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Something went wrong",
		})
	}

	rdbIncr := database.CreateClient(1)
	defer rdbIncr.Close()

	_ = rdbIncr.Incr(database.Ctx, "counter")

	return c.Redirect(val, fiber.StatusMovedPermanently)
}
