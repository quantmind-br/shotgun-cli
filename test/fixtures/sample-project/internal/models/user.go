package models

type User struct {
	ID    string
	Email string
}

func (u User) Valid() bool {
	return u.ID != "" && u.Email != ""
}
