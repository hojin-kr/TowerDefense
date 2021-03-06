package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gamejam/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"encoding/json"
)

var redisClient *redis.Client

const port string = ":8888"

func main() {
	// redis
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS"),
		Password: "",
		DB:       0,
	})
	// router
	r := gin.Default()
	r.LoadHTMLFiles("templates/index.html")
	r.Static("/static", "./static")
	r.GET("/leaderboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Leaderboard",
		})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// stage start
	r.POST("/stage/start", func(c *gin.Context) {
		var stage models.Stage
		if err := c.ShouldBind(&stage); err != nil {
			c.String(http.StatusBadRequest, "bad request")
			return
		}
		err := redisClient.ZIncrBy("stage:try", 1, strconv.Itoa(stage.ID)).Err()
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, true)
	})
	// stage clear
	r.POST("/stage/clear", func(c *gin.Context) {
		var stage models.Stage
		if err := c.ShouldBind(&stage); err != nil {
			c.String(http.StatusBadRequest, "bad request")
			return
		}
		log.Println(stage.ID)
		err := redisClient.ZIncrBy("stage:clear", 1, strconv.Itoa(stage.ID)).Err()
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, true)
	})
	// stage info get
	r.POST("/stage/get", func(c *gin.Context) {
		var stage models.Stage
		if err := c.ShouldBind(&stage); err != nil {
			c.String(http.StatusBadRequest, "bad request")
			return
		}
		try, err := redisClient.ZScore("stage:try", strconv.Itoa(stage.ID)).Result()
		if err != nil {
			try = 0
		}
		clear, err := redisClient.ZScore("stage:clear", strconv.Itoa(stage.ID)).Result()
		if err != nil {
			clear = 0
		}
		stage.TryCnt = try
		stage.ClearCnt = clear
		c.JSON(http.StatusOK, stage)
	})
	// stage all start count
	r.POST("/stage/get/startall", func(c *gin.Context) {
		tryStages, err := redisClient.ZRevRangeWithScores("stage:try", 0, -1).Result()
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, tryStages)
	})
	// stage all clear count
	r.POST("/stage/get/clearall", func(c *gin.Context) {
		clearStages, err := redisClient.ZRevRangeWithScores("stage:clear", 0, -1).Result()
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, clearStages)
	})
	// stage get all
	r.POST("/stage/get/all", func(c *gin.Context) {
		tryStages, err := redisClient.ZRevRangeWithScores("stage:try", 0, -1).Result()
		if err != nil {
			panic(err)
		}
		clearStages, err := redisClient.ZRevRangeWithScores("stage:clear", 0, -1).Result()
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"start": tryStages,
			"clear": clearStages,
		})
	})
	// stage info del
	r.POST("/stage/del", func(c *gin.Context) {
		err := redisClient.Del("stage:try").Err()
		if err != nil {
			panic(err)
		}
		err = redisClient.Del("stage:clear").Err()
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, true)
	})
	// set balance
	r.POST("/balance/set", func(c *gin.Context) {
		var data models.Balance
		if err := c.ShouldBind(&data); err != nil {
			c.String(http.StatusBadRequest, "bad request")
			return
		}
		err := redisClient.Set("balance", data.Data, 0).Err()
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, true)
	})
	// get balance
	r.POST("/balance/get", func(c *gin.Context) {
		data, err := redisClient.Get("balance").Result()
		if err != nil {
			panic(err)
		}
		res := models.Balances{}
		json.Unmarshal([]byte(data), &res)
		c.JSON(http.StatusOK, res)
	})
	r.Run(port) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
