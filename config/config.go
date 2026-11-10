package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

var ENV_LOCAL = "local"
var ENV_DEV = "dev"
var ENV_PROD = "prod"
var ENV string

var DB_URL string
var DB_NAME string = "imyr"
var JWT_KEY string

var AUTH_HASH_PASS_NONCE string
var AUTH_ADMIN_PASSWORD string

var STATIC_PATH = "/static"
var STATIC_DIR = "./static"
var PUBLIC_PATH = "/public"
var PUBLIC_DIR = "./public"
var STATIC_TEMPLATE_PATH = "/static/template"

var STATUS_DELETED = "deleted"
var STATUS_ACTIVE = "active"

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	ENV = os.Getenv("ENV")

	if ENV == "" {
		ENV = ENV_LOCAL
	}

	DB_URL = os.Getenv("MONGODB_URI")
	JWT_KEY = os.Getenv("JWT_KEY")

	AUTH_HASH_PASS_NONCE = os.Getenv("AUTH_HASH_PASS_NONCE")
	AUTH_ADMIN_PASSWORD = os.Getenv("AUTH_ADMIN_PASSWORD")
}
