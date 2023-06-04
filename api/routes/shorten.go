package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/hitesh22rana/blinkly/database"
	"github.com/hitesh22rana/blinkly/helpers"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

func ShortenURL(c *fiber.Ctx) error {
	var API_QUOTA string = os.Getenv("API_QUOTA")
	var API_DOMAIN string = os.Getenv("API_DOMAIN")

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

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	rdbId := database.CreateClient(0)
	defer rdbId.Close()

	val, _ = rdbId.Get(database.Ctx, id).Result()

	if val != "" {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Custom URL already exists",
		})
	}

	// Check if expiry is valid
	if body.Expiry < 0 || body.Expiry > 24*time.Hour {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid expiry",
		})
	}

	// Store URL in Redis
	err = rdbId.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Something went wrong",
		})
	}

	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	rdb.Decr(database.Ctx, IP)

	val, _ = rdb.Get(database.Ctx, IP).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := rdb.TTL(database.Ctx, IP).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = API_DOMAIN + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
