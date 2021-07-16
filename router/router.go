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
	uh := handlers.UserHandler{UserService: services.NewUserService(repo.NewUserRepoImpl())}

	app.Use(recover.New())
	api := app.Group("", logger.New())

	user := api.Group("application/storage/app/users")
	user.Get("/", middleware.IsLoggedIn, uh.GetAllUsers)
	user.Delete("/delete",middleware.IsLoggedIn,  uh.DeleteByID)
}

func Setup() *fiber.App {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		ExposeHeaders: "Authorization",
	}))

	SetupRoutes(app)
	return app
}
