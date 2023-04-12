package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

type emailRequest struct {
	To          []person `json:"to"`
	CC          []person `json:"cc"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	Attachments []string `json:"attachments"`
}
type person struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := gin.Default()
	router.LoadHTMLFiles("index.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.POST("/email", postEmail)

	router.Run("localhost:8080")
}

func postEmail(c *gin.Context) {
	// TODO: figure something else out xD
	token := strings.ToLower(c.Request.Header.Get("authorization"))
	if token != "bearer "+os.Getenv("MAIL_IT_TOKEN") {
		c.Status(http.StatusUnauthorized)
		return
	}
	var email emailRequest

	if err := c.BindJSON(&email); err != nil {
		c.IndentedJSON(http.StatusBadRequest, email)
		return
	}
	if email.To == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid To Address"})
		return
	}
	for _, to := range email.To {
		if to.Address == "" {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid To Address"})
			return
		}
	}
	for _, cc := range email.CC {
		if cc.Address == "" {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid CC Address"})
			return
		}
	}

	err := SendEmail(email)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Something happened, couldn't send email"})
		return
	}

	c.IndentedJSON(http.StatusCreated, email)
}

func SendEmail(email emailRequest) error {
	smtpHost := os.Getenv("MAIL_HOST")
	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	username := os.Getenv("MAIL_USERNAME")
	password := os.Getenv("MAIL_PASSWORD")
	fromName := os.Getenv("MAIL_FROM_NAME")
	fromEmail := os.Getenv("MAIL_FROM_EMAIL")

	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", fromEmail, fromName)

	for _, to := range email.To {
		m.SetAddressHeader("To", to.Address, to.Name)
	}
	for _, cc := range email.CC {
		m.SetAddressHeader("Cc", cc.Address, cc.Name)
	}
	m.SetHeader("Subject", email.Subject)
	m.SetBody("text/plain", email.Body)

	for _, attachment := range email.Attachments {
		// TODO: create directories per domain/subdirectory
		split := strings.Split(attachment, "/")
		filename := "images/" + split[len(split)-1]
		if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			err := DownloadFile(filename, attachment)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("Downloaded: " + attachment)
		}
		m.Attach(filename)
	}

	d := gomail.NewDialer(smtpHost, port, username, password)
	if err := d.DialAndSend(m); err != nil {
		log.Println(err)
		return err
	}
	return err
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
