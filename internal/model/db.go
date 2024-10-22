package model

type Product struct {
	ID          int64  `db:"id"          json:"id"`
	Name        string `db:"name"        json:"name"`
	Description string `db:"description" json:"description"`
	LastUser    string `db:"last_user"   json:"last_user"`
	UpdatedAt   string `db:"updated_at"  json:"updated_at"`
	CreatedAt   string `db:"created_at"  json:"created_at"`
}
