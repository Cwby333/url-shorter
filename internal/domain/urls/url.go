package urls

type URL struct {
	Id    int    `db:"id" json:"id"`
	URL   string `db:"url" json:"url"`
	Alias string `db:"alias" json:"alias"`
}
