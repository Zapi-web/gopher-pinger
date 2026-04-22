package keygen

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func GetKey() (ulid.ULID, error) {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)

	idMonotonic, err := ulid.New(ulid.Timestamp(t), entropy)
	if err != nil {
		return ulid.ULID{}, fmt.Errorf("failed to generate ulid")
	}

	return idMonotonic, nil
}
