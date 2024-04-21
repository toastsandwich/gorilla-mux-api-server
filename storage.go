package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
)

type Storage interface {
	CreateNewAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountByID(int) (*Account, error)
	GetAccounts() ([]*Account, error)
}

type MySqlStore struct {
	db *sql.DB
}

func NewMySqlStore() (*MySqlStore, error) {
	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "bank",
		AllowNativePasswords: true,
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &MySqlStore{
		db: db,
	}, nil
}

func (s *MySqlStore) Init() error {
	return s.CreateAccountTable()
}

func (s *MySqlStore) CreateAccountTable() error {
	query := `create table if not exists accounts (
		id int auto_increment primary key,
		account_number bigint not null,
		first_name varchar(255) not null,
		last_name varchar(255) not null,
		balance bigint unsigned not null,
		created_at longtext
	)`
	_, err := s.db.Exec(query)
	return err
}

func (s *MySqlStore) CreateNewAccount(acc *Account) error {
	query := `insert into accounts (
		account_number, first_name, last_name, balance, created_at
	) values (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, acc.Number, acc.FirstName, acc.LastName, acc.Balance, acc.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (s *MySqlStore) UpdateAccount(*Account) error {
	return nil
}

func (s *MySqlStore) DeleteAccount(id int) error {
	query := `delete from accounts where id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *MySqlStore) GetAccountByID(id int) (*Account, error) {
	query := `select * from accounts where id = ?`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccounts(rows)
	}
	return nil, fmt.Errorf("account %d not found", id)
}

func (s *MySqlStore) GetAccounts() ([]*Account, error) {
	query := `select * from accounts`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	accounts := []*Account{}
	for rows.Next() {

		account, err := scanIntoAccounts(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func scanIntoAccounts(rows *sql.Rows) (*Account, error) {
	account := new(Account)

	err := rows.Scan(
		&account.ID,
		&account.Number,
		&account.FirstName,
		&account.LastName,
		&account.Balance,
		&account.CreatedAt,
	)
	return account, err
}
