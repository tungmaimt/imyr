package auth

import (
	"maps"
	"time"

	"proj/imyr/config"

	"github.com/golang-jwt/jwt/v5"
)

func sign(data jwt.MapClaims) (s string, err error) {
	d := maps.Clone(data)
	secret := []byte(config.JWT_KEY)

	if v, ok := d["timestamps"]; ok == false || v == 0 {
		d["timestamps"] = time.Now().Unix()
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, d)
	s, err = t.SignedString(secret)
	return
}

func validate(s string) (d map[string]any, err error) {
	t, e := jwt.Parse(s, func(token *jwt.Token) (any, error) {
		return []byte(config.JWT_KEY), nil
	})

	if e != nil {
		err = e
		return
	}

	d = t.Claims.(jwt.MapClaims)
	return
}
