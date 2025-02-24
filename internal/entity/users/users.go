package users

type User struct {
	UUID        string `db:"uuid" json:"uuid"`
	Username    string `db:"username" json:"username"`
	Password    string `db:"password" json:"password"`
	Version     int    `db:"version" json:"version"`
	UserBlocked bool   `db:"user_blocked" json:"user_blocked"`
}
