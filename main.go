package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/alexedwards/argon2id"
	"github.com/joho/godotenv"
)

var ctx = context.TODO()
var database *mongo.Database
var users *mongo.Collection
var links *mongo.Collection
var CONNECTION_STRING string
var REGISTRATIONS_ENABLED bool
var BASE_URL string

var GENERAL_REGEX = `^[a-zA-Z0-9_-]{3,15}$`

type Link struct {
	ID      primitive.ObjectID `bson:"_id"`
	User    primitive.ObjectID `bson:"user"`
	Short   string             `bson:"short"`
	Link    string             `bson:"link"`
	Hits    int64              `bson:"hits"`
	Private bool               `bson:"private"`
}

type User struct {
	ID       primitive.ObjectID `bson:"_id"`
	Username string             `bson:"username"`
	Token    string             `bson:"token"`
}

func main() {
	godotenv.Load()
	CONNECTION_STRING = os.Getenv("CONNECTION_STRING")
	BASE_URL = os.Getenv("BASE_URL")
	REGISTRATIONS_ENABLED, _ = strconv.ParseBool(os.Getenv("REGISTRATIONS_ENABLED"))

	startupTime := time.Now()

	engine := html.New("./templates", ".html")
	engine.Reload(true)

	app := fiber.New(fiber.Config{Views: engine, ErrorHandler: func(c *fiber.Ctx, err error) error {
		var code int
		var message string

		if err != nil {
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
				message = http.StatusText(code)
			}

			if strings.HasPrefix(string(c.Request().URI().Path()), "/api") {
				return c.Status(code).JSON(fiber.Map{"error": message})
			}

			return c.Status(code).Render("error", fiber.Map{"Status": code, "Message": message})
		}

		return nil
	}})

	app.Static("/static", "./static")
	app.Use("/api/create", limiter.New(limiter.Config{Max: 5, Expiration: 1 * time.Minute}))
	app.Use("/api/join", limiter.New(limiter.Config{Max: 1, Expiration: 5 * time.Minute}))
	app.Use(logger.New(logger.Config{
		Format:     "${time} - ${status} - ${ip} ${method} ${path}\n",
		TimeFormat: "15:04:05 on 01/02/2006",
		TimeZone:   "America/Denver",
	}))

	app.Get("/api", func(c *fiber.Ctx) error {
		timeSinceStart := time.Now().Sub(startupTime)
		return c.JSON(
			fiber.Map{"ok": "ok", "uptime": timeSinceStart.Seconds(), "docs": "/api/docs"},
		)
	})

	app.Get("/api/docs", func(c *fiber.Ctx) error {
		return c.JSON(pages)
	})

	app.Post("/api/join", func(c *fiber.Ctx) error {
		if !REGISTRATIONS_ENABLED {
			return c.Status(fiber.StatusForbidden).
				JSON(fiber.Map{"error": "registrations have been disabled by the application owner"})
		}

		username := c.FormValue("username", "")
		if !testMatch(username, GENERAL_REGEX) {
			return c.Status(fiber.StatusBadRequest).
				JSON(fiber.Map{"error": "username does not match the regex: " + GENERAL_REGEX})
		}

		result := users.FindOne(ctx, bson.M{"username": username})
		if err := result.Err(); err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "username taken"})
		}

		token := generateToken(64)
		tokenHash, err := argon2id.CreateHash(token, argon2id.DefaultParams)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "an error when making your account"})
		}

		_, err = users.InsertOne(ctx, bson.D{
			{Key: "username", Value: username},
			{Key: "token", Value: tokenHash},
		})

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "an error occured when making your account"})
		}

		return c.JSON(
			fiber.Map{
				"ok":    "created account!",
				"info":  "please save this access token! you must use this for all future authenticated requests.",
				"token": token,
			},
		)
	})

	app.Post("/api/create", func(c *fiber.Ctx) error {
		token := c.Get("Authentication", "")
		username := c.FormValue("username", "")

		name := c.FormValue("name", generateToken(4))
		private := c.FormValue("private", "false")
		longLink := c.FormValue("link", "")

		if token == "" || username == "" {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"error": "unauthorized; username or token not provided"})
		}

		var user User
		result := users.FindOne(ctx, bson.M{"username": username})

		if result.Err() == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account not found"})
		}

		result.Decode(&user)

		compare, err := argon2id.ComparePasswordAndHash(token, user.Token)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "an error occured"})
		}

		if !compare {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "incorrect token"})
		}

		if (private != "false" && private != "true") || longLink == "" {
			return c.Status(fiber.StatusBadRequest).
				JSON(fiber.Map{"error": "link privacy must be true or false"})
		}

		otherLink := links.FindOne(ctx, bson.M{"short": name})
		if otherLink.Err() != mongo.ErrNoDocuments {
			return c.Status(fiber.StatusConflict).
				JSON(fiber.Map{"error": "link name already exists"})
		}

		_, err = url.ParseRequestURI(longLink)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad link"})
		}

		if !testMatch(name, GENERAL_REGEX) {
			return c.Status(fiber.StatusBadRequest).
				JSON(fiber.Map{"error": "shortened name does not match the regex: " + GENERAL_REGEX})
		}

		privateBool, _ := strconv.ParseBool(private)

		_, err = links.InsertOne(ctx, bson.D{
			{Key: "user", Value: user.ID},
			{Key: "short", Value: name},
			{Key: "link", Value: longLink},
			{Key: "hits", Value: 0},
			{Key: "private", Value: privateBool},
		})

		return c.JSON(
			fiber.Map{
				"ok":      "created link",
				"link":    BASE_URL + "/l/" + name,
				"name":    name,
				"long":    longLink,
				"private": private,
			},
		)
	})

	app.Get("/l/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")

		var link Link
		result := links.FindOne(ctx, bson.M{"short": name})

		if result.Err() == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "link not found"})
		}

		result.Decode(&link)

		links.UpdateOne(ctx, bson.M{"short": name}, bson.M{"$inc": bson.M{"hits": 1}})

		return c.Redirect(link.Link, fiber.StatusTemporaryRedirect)
	})

	app.Get("/api/link/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		token := c.Get("Authentication", "")
		username := c.FormValue("username", "")

		var link Link
		result := links.FindOne(ctx, bson.M{"short": name})

		if result.Err() == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "link not found"})
		}

		result.Decode(&link)

		if link.Private {
			if token == "" || username == "" {
				return c.Status(fiber.StatusUnauthorized).
					JSON(fiber.Map{"error": "unauthorized; username or token not provided"})
			}

			var user User
			result := users.FindOne(ctx, bson.M{"username": username})

			if result.Err() == mongo.ErrNoDocuments {
				return c.Status(fiber.StatusUnauthorized).
					JSON(fiber.Map{"error": "account not found"})
			}

			result.Decode(&user)

			compare, err := argon2id.ComparePasswordAndHash(token, user.Token)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).
					JSON(fiber.Map{"error": "an error occured"})
			}

			if !compare {
				return c.Status(fiber.StatusUnauthorized).
					JSON(fiber.Map{"error": "incorrect token"})
			}

			return c.JSON(fiber.Map{
				"name":    link.Short,
				"long":    link.Link,
				"short":   BASE_URL + "/l/" + link.Short,
				"hits":    link.Hits,
				"private": true,
			})
		}

		return c.JSON(fiber.Map{
			"name":  link.Short,
			"long":  link.Link,
			"short": BASE_URL + "/l/" + link.Short,
			"hits":  link.Hits,
		})
	})

	app.Get("/api/links/:username", func(c *fiber.Ctx) error {
		lookupUser := c.Params("username")

		token := c.Get("Authentication", "")
		username := c.FormValue("username", "")

		var user User
		userLookup := users.FindOne(ctx, bson.M{"username": lookupUser})

		if userLookup.Err() == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).
				JSON(fiber.Map{"error": "user not found"})
		}

		userLookup.Decode(&user)

		authenticated := false
		if token != "" && username != "" && user.Username == username {
			compare, _ := argon2id.ComparePasswordAndHash(token, user.Token)
			if !compare {
				return c.Status(fiber.StatusUnauthorized).
					JSON(fiber.Map{"error": "incorrect token"})
			}

			authenticated = true
		}

		var linkList []interface{}
		linkLookup, err := links.Find(ctx, bson.M{"user": user.ID})

		if err != nil || (linkLookup.Err() != nil && linkLookup.Err() != mongo.ErrNoDocuments) {
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "an error occured when fetching links"})
		}

		for linkLookup.Next(ctx) {
			var link Link
			linkLookup.Decode(&link)

			if !authenticated && link.Private {
				continue
			}

			linkList = append(linkList, fiber.Map{
				"name":  link.Short,
				"long":  link.Link,
				"short": BASE_URL + "/l/" + link.Short,
				"hits":  link.Hits,
			})
		}

		if len(linkList) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no links found"})
		}

		return c.JSON(linkList)
	})

	connect()
	app.Listen(":3000")
}

func connect() {
	clientOptions := options.Client().
		ApplyURI(CONNECTION_STRING)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	database = client.Database("goto-raspapi")

	users = database.Collection("users")
	links = database.Collection("links")

	fmt.Println("Connected to MongoDB!")
}
