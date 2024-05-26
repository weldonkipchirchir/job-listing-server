package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weldonkipchirchir/job-listing-server/db"
	"github.com/weldonkipchirchir/job-listing-server/routes"
)

func main() {
	router := gin.Default()

	err := db.DbConnection()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	routes.SetUpusers(router)
	routes.JobRoutes(router)

	//create server
	serv := &http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	//start the server
	go func() {
		log.Println("Server is running on http://localhost:8000")
		if err := serv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server forced to shutdown")
		}
	}()

	//handlesignals to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("Server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := serv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", err)
	}

	db.DbDisconnect()

	log.Println("Server exiting")
}
