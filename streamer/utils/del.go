package utils

import "github.com/gofrs/uuid"

func Del[T any](m map[uuid.UUID]T, id uuid.UUID) {
	var _, ok = m[id]
	if ok {
		delete(m, id)
	}
}
