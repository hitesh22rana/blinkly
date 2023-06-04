package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/hitesh22rana/blinkly/database"
	"github.com/hitesh22rana/blinkly/helpers"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

var API_QUOTA string = os.Getenv("API_QUOTA")

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	IP := c.IP()

	// Rate Limiting
	rdb := database.CreateClient(1)
	defer rdb.Close()

	val, err := rdb.Get(database.Ctx, IP).Result()

	if err == redis.Nil {
		_ = rdb.Set(database.Ctx, IP, API_QUOTA, 30*60*time.Second).Err()
	} else {
		val, _ = rdb.Get(database.Ctx, IP).Result()
		valInt, _ := strconv.Atoi(val)

		if valInt <= 0 {
			limit, _ := rdb.TTL(database.Ctx, IP).Result()
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":            "Rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	// Check if URL is valid
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	// Check for domain errors
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "URL is not allowed to be shortened",
		})
	}

	// Enforce HTTP
	body.URL = helpers.EnforceHTTP(body.URL)

	rdb.Decr(database.Ctx, IP)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "URL shortened successfully",
		"data":    body,
	})
}
