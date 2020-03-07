package utils

import (
	"errors"
	"log"
	"os"
)

var (
	ErrBadArrayLen        = errors.New("bad array len")
	ErrBadArrayLenTooLong = errors.New("bad array len, too long")

	ErrBadBulkBytesLen        = errors.New("bad bulk bytes len")
	ErrBadBulkBytesLenTooLong = errors.New("bad bulk bytes len, too long")

	ErrBadMultiBulkLen     = errors.New("bad multi-bulk len")
	ErrBadMultiBulkContent = errors.New("bad multi-bulk content, should be bulkbytes")
)

func ErrorNew(msg string) error {
	return errors.New("error occur, msg ")
}

func ErrorsTrace(err error) error {
	if err != nil {
		log.Println("errors Tracing", err.Error())
	}
	return err
}

func CheckError(err error) {
	if err != nil {
		log.Println("err ", err.Error())
		os.Exit(1)
	}
}