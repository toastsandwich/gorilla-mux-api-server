package main

import (
	"math/rand"
	"time"
)

type APIError struct {
	Error string
}

type TransferRequest struct {
	ToAccount int64 `json:"toAccount"`
	Amount    int   `json:"amount"`
}

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Account struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Number    int64  `json:"account_number"`
	Balance   uint64 `json:"balance"`
	CreatedAt string `json:"created_at"`
}

func NewAccount(fn, ln string) *Account {
	return &Account{
		FirstName: fn,
		LastName:  ln,
		Number:    rand.Int63n(10000),
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
}
