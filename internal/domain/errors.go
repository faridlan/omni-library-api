package domain

import "errors"

var (
	ErrInternalServerError = errors.New("terjadi kesalahan pada server")
	ErrNotFound            = errors.New("data tidak ditemukan")
	ErrConflict            = errors.New("data sudah ada (konflik)")
	ErrBadParamInput       = errors.New("parameter atau format data tidak valid")
	ErrLimitExceeded       = errors.New("kuota API eksternal habis")
)
