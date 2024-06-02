package auth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPasswordHash(t *testing.T) {
	password := "Zhandar1503"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hash)

	valid := CheckPassword(password, hash)

	assert.True(t, valid, true)
}
