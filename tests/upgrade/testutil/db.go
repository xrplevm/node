package testutil

import (
	"io"
	"os"
)

func CopyNodeDB(from, to string) error {
	fromDB, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromDB.Close()

	toDB, err := os.Create(to)
	if err != nil {
		return err
	}
	defer toDB.Close()

	_, err = io.Copy(toDB, fromDB)
	return err
}
