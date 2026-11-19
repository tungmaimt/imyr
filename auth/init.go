package auth

import (
	"context"
	"fmt"
	"proj/imyr/config"
	"proj/imyr/db"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func init() {
	fmt.Println("INIT AUTH")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	admin, err := db.FindOne[UserModel](ctx, map[string]any{"username": "tungmai"}, nil)

	if err == nil {
		fmt.Println(admin.ID, admin.Username, admin.Email, admin.Created.String(), "admin exists")
		return
	}

	if err != mongo.ErrNoDocuments {
		fmt.Println(err, "Error init auth")
		return
	}

	fmt.Println(err, "error find admin")

	doc := &UserModel{
		Email:    "tungmaimtt@gmail.com",
		Username: "tungmai",
		HashPass: hashPass(config.AUTH_ADMIN_PASSWORD),
		Role:     ROLE_ADMIN,
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	ir, err := db.InsertOne(ctx, doc, nil)

	if err != nil || ir == nil {
		fmt.Println(err, ir, "Error create user init auth")
		return
	}

	fmt.Println(ir, "Create admin auth init")
}
