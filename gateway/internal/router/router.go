package router

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/avran02/fileshare/gateway/internal/controller"
	customMiddleware "github.com/avran02/fileshare/gateway/internal/middlaware"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	chi.Router
	controllers *controller.Controllers
}

func (router *Router) getUserRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", router.controllers.UsersController.Login)
	r.Post("/register", router.controllers.UsersController.Register)
	r.Post("/refresh-token", router.controllers.UsersController.RefreshToken)
	r.Post("/logout", router.controllers.UsersController.Logout)

	return r
}

func (router *Router) getFilesRoutes() chi.Router {
	r := chi.NewRouter()
	authMiddlaware := customMiddleware.GetAuthMiddleware(router.controllers.UsersController.GetGrpcClient())
	r.Use(authMiddlaware)

	r.Post("/upload", router.controllers.FilesController.Upload)
	r.Get("/download", router.controllers.FilesController.Download)
	r.Delete("/rm", router.controllers.FilesController.Rm)
	r.Get("/ls", router.controllers.FilesController.Ls)
	return r
}

func (router *Router) getShareRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/share", router.controllers.ShareController.Share)
	r.Delete("/unshare", router.controllers.ShareController.Unshare)

	return r
}

func New(controllers controller.Controllers) *Router {
	router := &Router{
		controllers: &controllers,
		Router:      chi.NewRouter(),
	}

	router.Router.Use(middleware.Recoverer)
	router.Router.Use(middleware.Logger)

	router.Router.Route("/api/v1", func(r chi.Router) {
		r.Mount("/user", router.getUserRoutes())
		r.Mount("/files", router.getFilesRoutes())
		r.Mount("/share", router.getShareRoutes())
	})

	printRoutes(router.Router)
	return router
}

func printRoutes(router chi.Routes) {
	slog.Info("Routes:")
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		loggingStr := fmt.Sprintf("Method: %s, Route: %s", method, route)
		slog.Info(loggingStr)
		return nil
	}
	err := chi.Walk(router, walkFunc)
	if err != nil {
		slog.Error(err.Error())
	}
}
