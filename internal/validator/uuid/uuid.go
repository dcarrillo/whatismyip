package uuid

import (
	"github.com/google/uuid"
)

func IsValid(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
