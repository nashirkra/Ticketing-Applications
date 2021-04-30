package main

import (
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jasonlvhit/gocron"
	"github.com/nashirkra/Ticketing-Applications/conf"
	"github.com/nashirkra/Ticketing-Applications/controller"
	"github.com/nashirkra/Ticketing-Applications/middleware"
	"github.com/nashirkra/Ticketing-Applications/repository"
	"github.com/nashirkra/Ticketing-Applications/service"
	"gorm.io/gorm"
)

var (
	// setup Database Connection
	db        *gorm.DB                         = conf.SetupDBConn()
	userRepo  repository.UserRepository        = repository.NewUserRepository(db)
	eventRepo repository.EventRepository       = repository.NewEventRepository()
	trxRepo   repository.TransactionRepository = repository.NewTransactionRepository()
	//setup services
	jwtServ  service.JWTService         = service.NewJWTService()
	authServ service.AuthService        = service.NewAuthService(userRepo)
	userServ service.UserService        = service.NewUserService(userRepo)
	evServ   service.EventService       = service.NewEventService(eventRepo)
	trxServ  service.TransactionService = service.NewTransactionService(trxRepo, userRepo, eventRepo)
	//setup controller
	authController  controller.AuthController        = controller.NewAuthController(authServ, jwtServ)
	userController  controller.UserController        = controller.NewUserController(userServ, jwtServ)
	eventController controller.EventController       = controller.NewEventController(evServ, jwtServ)
	trxController   controller.TransactionController = controller.NewTransactionController(trxServ, userServ, jwtServ)
	payController   controller.PaymentController     = controller.NewPaymentController(trxServ, jwtServ)
)

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d%02d/%02d", year, month, day)
}

func startCron() {
	context := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// sample list of mail
	client.Del(context, "dailyMail")
	client.LPush(context, "dailyMail", "This is Mail 1", "This is Mail 2", "This is Mail 3", "This is Mail 4", "This is Mail 5")

	s := gocron.NewScheduler()
	// s.Every(1).Day().At("09:00").Do(service.DailyMail)
	s.Every(3).Seconds().Do(service.DailyMail)
	<-s.Start()
}

func main() {
	go startCron()
	r := gin.Default()

	r.SetFuncMap(template.FuncMap{
		"formatAsDate": formatAsDate,
	})
	// AuthRoutes := r.Group("api/auth", middleware.AuthorizeJWT(jwtServ))
	authRoutes := r.Group("api/auth")
	{
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/register", authController.Register)
	}

	// User Routes
	userRoutes := r.Group("api/user", middleware.AuthorizeJWT(jwtServ))
	{
		userRoutes.PUT("/profile", userController.Update)
		userRoutes.GET("/profile", userController.Profile)
	}

	// Admin Routes
	adminRoutes := r.Group("api/admin", middleware.AuthorizeJWT(jwtServ))
	{
		adminRoutes.PUT("/profile", userController.Update)
		adminRoutes.GET("/profile", userController.Profile)
		adminRoutes.GET("/users", userController.All)
		// Event create handler
		adminRoutes.POST("/event", eventController.Create)

	}

	eventRoutes := r.Group("api/event", middleware.AuthorizeJWT(jwtServ))
	{
		// Event create handler
		eventRoutes.GET("/:id", eventController.GetEvent)
		eventRoutes.GET("/all", eventController.AllEvent)
		eventRoutes.Use(middleware.LogResponseBody)
		eventRoutes.POST("/:id", trxController.Create)
	}

	paymentRoutes := r.Group("api/payment")
	{
		paymentRoutes.GET("/:token", payController.InfoPayment)
		paymentRoutes.POST("/:token", payController.Pay)
		paymentRoutes.POST("/cancel/:token", payController.Cancel)
	}

	r.Run()
}
