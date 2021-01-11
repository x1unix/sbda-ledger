package user

import "github.com/jackc/pgtype"

type GroupID = pgtype.UUID
type Groups = []Group

type GroupInfo struct {
	Name    string `json:"name" db:"id"`
	OwnerID ID     `json:"owner_id" db:"owner_id"`
}

type Group struct {
	GroupInfo
	ID pgtype.UUID `json:"id" db:"id"`
}
