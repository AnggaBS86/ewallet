package service

import "errors"

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var ErrUnauthorized = errors.New("unauthorized")
var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrSelfTransfer = errors.New("self transfer")
var ErrReceiverNotFound = errors.New("receiver not found")
var ErrWalletNotFound = errors.New("wallet not found")
