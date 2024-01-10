package main

import (
	"errors"

	"github.com/kafkago/db"
	"github.com/kafkago/db/models"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
)


func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})

	messageService := db.NewRedis[models.Message](client)
	http := echo.New()

	http.POST("/message", func(ctx echo.Context) error {
		content := ctx.Request().PostFormValue("message")
		message := models.Message{Message: content}
		err := messageService.Save(ctx.Request().Context(), message)
		if err != nil {
			return ctx.String(500, err.Error())
		}
		return ctx.String(201, message.UID)
	})

	http.GET("/message/:uid", func(ctx echo.Context) error {
		uid := ctx.Param("uid")
		message, err := messageService.Get(ctx.Request().Context(), uid)
		if errors.Is(err, redis.Nil) {
			return ctx.String(404, "message not found")
		} else if err != nil {
			return ctx.String(500, err.Error())
		}
		return ctx.String(200, message.Message)
	})
	http.Logger.Fatal(http.Start(":1234"))
	
}