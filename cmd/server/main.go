package main

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	"github.com/vterdunov/go-microservice-fiber/internal/handler"
	"github.com/vterdunov/go-microservice-fiber/internal/storage"
)

func main() {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Recovery middleware
	app.Use(recover.New())

	// Logger middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	// Prometheus metrics
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by status code, method and path",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latencies in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	httpErrorsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors (status >= 400)",
		},
		[]string{"method", "path", "status"},
	)

	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, httpErrorsTotal)

	// Prometheus metrics middleware
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := c.Response().StatusCode()
		statusStr := strconv.Itoa(status)
		method := c.Method()
		path := c.Route().Path
		if path == "" {
			path = c.Path()
		}

		httpRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
		httpRequestDuration.WithLabelValues(method, path, statusStr).Observe(duration)

		if status >= 400 {
			httpErrorsTotal.WithLabelValues(method, path, statusStr).Inc()
		}

		return err
	})

	// Metrics endpoint
	app.Get("/metrics", func(c *fiber.Ctx) error {
		handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
		handler(c.Context())
		return nil
	})

	// Rate limiter: 1000 requests per second
	app.Use(limiter.New(limiter.Config{
		Max:        100000,
		Expiration: 1 * time.Second,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded",
			})
		},
	}))

	// Storage
	store := storage.NewMemoryStorage()

	// Handlers
	userHandler := handler.NewUserHandler(store)

	// Routes
	api := app.Group("/api")
	users := api.Group("/users")

	users.Get("/", userHandler.GetAll)
	users.Get("/:id", userHandler.Get)
	users.Post("/", userHandler.Create)
	users.Put("/:id", userHandler.Update)
	users.Delete("/:id", userHandler.Delete)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	log.Fatal(app.Listen(":3000"))
}
