package auth

import (
	"context"
	"database/sql"


	"github.com/volatiletech/authboss/v3"
)




type User struct {
	ID string
	Email    string
	Password string
}

func (u *User) GetPID() string           { return u.ID }
func (u *User) GetPassword() string      { return u.Password }
func (u *User) GetEmail() string         { return u.Email }
func (u *User) GetConfirmed() bool       { return true }
func (u *User) GetLocked() bool          { return false }
func (u *User) GetRecoverSelector() string { return "" }
func (u *User) GetRecoverToken() string    { return "" }
func (u *User) GetConfirmSelector() string { return "" }
func (u *User) GetConfirmToken() string    { return "" }

// === Setters ("Put") to allow mutation ===

func (u *User) PutPID(pid string)                { u.ID = pid }
func (u *User) PutPassword(password string)      { u.Password = password }
func (u *User) PutEmail(email string)            { u.Email = email }
func (u *User) PutConfirmed(confirmed bool)      {}
func (u *User) PutLocked(locked bool)            {}
func (u *User) PutRecoverSelector(selector string) {}
func (u *User) PutRecoverToken(token string)        {}
func (u *User) PutConfirmSelector(selector string)  {}
func (u *User) PutConfirmToken(token string)        {}

type UserStorer struct {
	DB *sql.DB
}

func (s UserStorer) Load(ctx context.Context, key string) (authboss.User, error) {
	row := s.DB.QueryRowContext(ctx, `SELECT email, password FROM users WHERE email=$1`, key)
	var u User
	if err := row.Scan(&u.Email, &u.Password); err != nil {
		return nil, authboss.ErrUserNotFound
	}
	return &u, nil
}

func (s UserStorer) Save(ctx context.Context, user authboss.User) error {
	u := user.(*User)
	_, err := s.DB.ExecContext(ctx, `INSERT INTO users (email, password) VALUES ($1, $2) 
		ON CONFLICT (email) DO UPDATE SET password = $2`, u.Email, u.Password)
	return err
}
