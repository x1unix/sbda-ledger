package user

import "github.com/jackc/pgtype"

type GroupID = pgtype.UUID
type Groups = []Group

type Group struct {
	ID      pgtype.UUID `json:"id" db:"id"`
	Name    string      `json:"name" db:"name"`
	OwnerID ID          `json:"owner_id" db:"owner_id"`
}

type GroupInfo struct {
	Group
	Members Users `json:"members"`
}
