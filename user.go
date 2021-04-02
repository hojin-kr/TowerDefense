package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-redis/redis/v8"
	"context"
	"strconv"
)

const port string = ":8888"

var db *sql.DB
var ctx = context.Background()
// User basic info
type User struct {
	ID         int    `json:"id"`
	PlatformID string `json:"platform_id"`
	Platform   string `json:"platform"`
	DeviceID  string `json:"device_id"`
}
// Rank rank info
type Rank struct {
	ID int `json:"id"`
	Score float64 `json:"score"`
	Rank int64 `json:"rank"`
}

// Login user return game id
func Login(c *gin.Context) {
	var user User
	if err := c.ShouldBind(&user); err != nil {
		c.String(http.StatusBadRequest, "bad request")
		return
	}
	row := db.QueryRow("SELECT id, platform_id, platform, device_id from user where platform_id = ? AND platform = ?", user.PlatformID, user.Platform)
	row.Scan(&user.ID, &user.PlatformID, &user.Platform, &user.DeviceID)
	if user.ID == 0 {
		// sign up
		rs, err := db.Exec("INSERT INTO user(platform_id, platform, device_id) VALUES (?, ?, ?)", user.PlatformID, user.Platform, user.DeviceID)
		if err != nil {
			log.Fatalln(err)
		}
		id, err := rs.LastInsertId()
		if err != nil {
			log.Fatalln(err)
		}
		user.ID = int(id)
	}
	c.JSON(http.StatusOK, user)
}
// ChangePlatfrom change platform
func ChangePlatfrom(c *gin.Context)  {
	var user User
	if err := c.ShouldBind(&user); err != nil {
		c.String(http.StatusBadRequest, "bad request")
		return
	}
	if user.ID != 0 {
		rs, err := db.Exec("UPDATE user SET platform_id = ?, platform = ? WHERE id = ? AND device_id = ?", user.PlatformID, user.Platform, user.ID, user.DeviceID)
		if err != nil || rs == nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
	}
	c.JSON(http.StatusOK, user)
}
// checkDeviceID 입력받은 디바이스 아이디가 유효한지 확인
func checkDeviceID(ID int, DeviceID string) (bool) {
	var _DeviceID string
	row := db.QueryRow("SELECT device_id from user where id = ?", ID)
	row.Scan(&_DeviceID)
	if(_DeviceID != DeviceID) {
		return false
	}
	return true
}
func incrScore(c *gin.Context) {
	var rank Rank
	if err := c.ShouldBind(&rank); err != nil {
		c.String(http.StatusBadRequest, "bad request")
		return
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:	"redis:6379",
		Password: "",
		DB:	0,
	})
	err := rdb.ZIncrBy(ctx, "rank", rank.Score, strconv.Itoa(rank.ID)).Err()
	if err != nil {
		panic(err)
	}
	userRank, err := rdb.ZRevRank(ctx, "rank", strconv.Itoa(rank.ID)).Result()
	if err != nil {
		panic(err)
	}
	rank.Rank = userRank
	c.JSON(http.StatusOK, rank)
}
// getLeaderboard getLeaderboard
func getLeaderboard(c *gin.Context) {
	rdb := redis.NewClient(&redis.Options{
		Addr:	"redis:6379",
		Password: "",
		DB:	0,
	})
	ranks, err := rdb.ZRevRangeWithScores(ctx, "rank", 0, 500).Result()
	if err != nil {
		panic(err)
	}
	log.Println(ranks)
	c.JSON(http.StatusOK, ranks)
}

func main() {
	// database
	var err error
	db, err = sql.Open("mysql", "app:1q2w3e4r@tcp(rdb:3306)/user")
	if err != nil {
		log.Println(err)
	}
	// main() 종료되면 디비 연결 헤제 / Then finished main close db connet
	defer db.Close()
	//router
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
	// sign up or sign in return game id
	r.POST("/login", Login)
	r.POST("/changeplatfrom", ChangePlatfrom)
	r.POST("/incrscore", incrScore)
	r.POST("/getleaderboard", getLeaderboard)
	r.Run(port) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
