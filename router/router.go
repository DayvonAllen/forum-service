package router

import (
	"example.com/app/handlers"
	"example.com/app/middleware"
	"example.com/app/repo"
	"example.com/app/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupRoutes(app *fiber.App) {
	th := handlers.ThreadHandler{ThreadService: services.NewThreadService(repo.NewThreadRepoImpl())}

	app.Use(recover.New())
	api := app.Group("", logger.New())

	threads := api.Group("/threads")
	threads.Get("/", middleware.IsLoggedIn, th.GetAllThreads)
	threads.Post("/",  th.CreateThread)
	threads.Delete("/delete", th.DeleteByID)
}

func Setup() *fiber.App {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		ExposeHeaders: "Authorization",
	}))

	SetupRoutes(app)
	return app
}
