package router

import (
	"example.com/app/handlers"
	"example.com/app/repo"
	"example.com/app/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupRoutes(app *fiber.App) {
	th := handlers.ThreadHandler{ThreadService: services.NewThreadService(repo.NewThreadRepoImpl())}
	reh := handlers.ReplyHandler{ReplyService: services.NewReplyService(repo.NewReplyRepoImpl())}
	ph := handlers.PostHandler{PostService: services.NewPostService(repo.NewPostRepoImpl())}

	app.Use(recover.New())
	api := app.Group("", logger.New())

	threads := api.Group("/threads")
	threads.Get("/name", th.GetThreadByName)
	threads.Get("/", th.GetAllThreads)
	threads.Post("/",  th.CreateThread)
	threads.Delete("/delete", th.DeleteByID)

	posts := api.Group("/post")
	posts.Post("/:id", ph.CreatePostOnThread)
	posts.Put("/like/:id", ph.LikePost)
	posts.Put("/dislike/:id", ph.DisLikePost)
	posts.Put("/:id", ph.UpdateById)
	posts.Delete("/:id", ph.DeleteById)

	reply := api.Group("/post/reply")
	reply.Post("/:id", reh.CreateReply)
	reply.Put("/like/:id", reh.LikeReply)
	reply.Put("/dislike/:id", reh.DisLikeReply)
	reply.Put("/:id", reh.UpdateById)
	reply.Delete("/:id", reh.DeleteById)
}

func Setup() *fiber.App {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		ExposeHeaders: "Authorization",
	}))

	SetupRoutes(app)
	return app
}
