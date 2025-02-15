package auth

type User struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Pass  string `db:"pass"`
	Coins int    `db:"coins"`
}
