package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PunMung-66/todoAPI/auth"
	"github.com/PunMung-66/todoAPI/todo"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	// "github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// This is set by build flags, we can use it to show build info in the future
// use: go build -ldflags "-X main.buildcommit=$(git rev-parse --short HEAD) -X main.buildtime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o app
var (
	buildcommit = "dev"
	buildtime   = time.Now().String()
)

func main() {

	// make a liveness probe for kubernetes
	_, err := os.Create("/tmp/live")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove("/tmp/live")

	err = godotenv.Load("local.env")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	sign := os.Getenv("SIGN")

	if err != nil {
		log.Printf("please consider environment variable: %s\n", err)
	}

	// This line is for sqlite, we can change it to mysql or postgresql later
	// db, err := gorm.Open(sqlite.Open(os.Getenv("DB_CONN")), &gorm.Config{})

	// This line is for mysql, we can change it to postgresql later
	db, err := gorm.Open(mysql.Open(os.Getenv("DB_CONN")), &gorm.Config{})
	if err != nil {
		// show reason for error
		panic("failed to connect database: " + err.Error())
	}

	db.AutoMigrate(&todo.Todo{})

	r := gin.Default()
	// set cors
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:8080",
	}

	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Transaction-ID"}

	r.Use(cors.New(config))

	// readiness probe for kubernetes, we can use it to show build info in the future
	r.GET("/health", func(c *gin.Context) {
		c.Status(200)
	})

	// limit handler for testing, we can remove it later
	r.GET("/limit", limitedHandler)

	// This is for testing build info, we can remove it later
	r.GET("/x", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"buildcommit": buildcommit,
			"buildtime":   buildtime,
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	protected := r.Group("")
	protected.Use(auth.Protect([]byte(os.Getenv(sign))))

	r.GET("/token", auth.AccessToken([]byte(os.Getenv(sign))))

	todoHandler := todo.NewTodoHandler(db)
	protected.POST("/todos", todoHandler.NewTask)
	protected.GET("/todos", todoHandler.List)
	protected.DELETE("/todos/:id", todoHandler.Delete)

	// r.Run(":" + port) // if we wnt to do gracefull shoutdown we need to use net http ,so can not use Gin

	// This part is for graceful shutdown, we need to use net http instead of Gin's Run method
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s := &http.Server{
		Addr:           ":" + os.Getenv("PORT"),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	stop()
	fmt.Println("Shutting down gracefully, press Ctrl + C again to force")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}
}

// do a late limit

var limiter = rate.NewLimiter(5, 5) // 1 request per second with a burst of 5

// add handler to route to test
// use echo GET http://localhost:8081/limit | vegeta attack -rate=10/s -duration=1s | vegeta report
// to run in commad prompt and see the result
func limitedHandler(c *gin.Context) {
	// if limit is exceeded, return 429 Too Many Requests
	if !limiter.Allow() {
		c.AbortWithStatus(http.StatusTooManyRequests)
		return
	}

	c.JSON(200, gin.H{
		"message" : "pong",
	})
}

