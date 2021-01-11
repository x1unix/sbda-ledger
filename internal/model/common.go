package model

import (
	"github.com/jackc/pgtype"
	"github.com/x1unix/sbda-ledger/internal/web"
)

// DecodeUUID decodes UUID from string
func DecodeUUID(str string) (*pgtype.UUID, error) {
	if str == "" {
		return nil, web.NewErrBadRequest("empty resource id")
	}

	id := new(pgtype.UUID)
	err := id.DecodeText(nil, []byte(str))
	if err != nil {
		return nil, web.NewErrBadRequest("invalid resource id: %s", err)
	}
	return id, nil
}

// DecodeUUIDs decodes multiple uuids from string
// and returns slice of decoded ids.
func DecodeUUIDs(strs ...string) ([]pgtype.UUID, error) {
	out := make([]pgtype.UUID, 0, len(strs))
	for _, str := range strs {
		id, err := DecodeUUID(str)
		if err != nil {
			return nil, err
		}
		out = append(out, *id)
	}
	return out, nil
}
