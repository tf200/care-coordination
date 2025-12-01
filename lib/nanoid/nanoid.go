package nanoid

import gonanoid "github.com/matoous/go-nanoid/v2"

const (
	size  = 20
	chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz-"
)

func Generate() string {
	id, err := gonanoid.Generate(chars, size)
	if err != nil {
		panic(err)
	}
	return id
}
