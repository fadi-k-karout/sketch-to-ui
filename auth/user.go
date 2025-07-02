package auth

import (
	"fmt"
	"image/png"

	"os"
	"path/filepath"
	"strconv"
	"unicode/utf8"

	"github.com/disintegration/letteravatar"
)

type ID int

type Role string

const (
	Regular Role = "member"

	Admin Role = "admin"
)

func (id ID) String() string {
	return strconv.Itoa(int(id))
}

func ParseUserID(s string) (ID, error) {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return ID(id), nil
}

type User struct {
	ID        ID     `db:"id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
	Password  string `db:"password"`
	AvatarURI string `db:"avatar_uri"`
	Role      Role  `db:"user_role"`
}

func (u *User) generateAvatar() (string, error) {
	firstLetter, _ := utf8.DecodeRuneInString(u.FirstName)

	img, err := letteravatar.Draw(75, firstLetter, nil)
	if err != nil {
		return "", err
	}

	// Create the avatars directory if it doesn't exist
	avatarsDir := "./static/images/avatars"
	err = os.MkdirAll(avatarsDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	// Save the image to the avatars directory
	filename := fmt.Sprintf("%s-%s.png", u.FirstName, u.ID.String())
	filepath := filepath.Join(avatarsDir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	err = png.Encode(file, img)
	if err != nil {
		return "", err
	}

	return filepath, nil
}
