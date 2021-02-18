package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

const port string = ":8888"

var db *sql.DB

// User basic info
type User struct {
	ID         int    `json:"id"`
	PlatformID string `json:"platform_id" binding:"required"`
	Platform   string `json:"platform" binding:"required"`
	DeviceID  string `json:"device_id"`
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
		// 검증
		if(!checkDeviceID(user.ID, user.DeviceID)) {
			c.JSON(http.StatusBadRequest, user)
		}
		rs, err := db.Exec("UPDATE TABLE user SET platfrom_id = ?, platfrom = ? WHERE id = ?", user.PlatformID, user.Platform, user.ID)
		if err != nil || rs == nil {
			log.Fatalln(err)
		}
	}
	c.JSON(http.StatusOK, user)
}

func checkDeviceID(ID int, DeviceID string) (bool) {
	var _DeviceID string
	row := db.QueryRow("SELECT device_id from user where id = ?", ID)
	row.Scan(&_DeviceID)
	if(_DeviceID != DeviceID) {
		return false
	}
	return true
}

func main() {
	// database
	var err error
	db, err = sql.Open("mysql", "USER:PASWORD@tcp(127.0.0.1:3306)/user")
	if err != nil {
		log.Println(err)
	}
	// main() 종료되면 디비 연결 헤제 / Then finished main close db connet
	defer db.Close()
	//router
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// sign up or sign in return game id
	r.POST("/login", Login)
	r.POST("/changeplatfrom", ChangePlatfrom)
	r.Run(port) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}