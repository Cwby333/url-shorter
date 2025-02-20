package users

type User struct {
	UUID     string `db:"uuid" json:"uuid"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
}
