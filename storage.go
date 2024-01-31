package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountById(int) (*Account, error)
	GetAccountByJWT(string) (*Account, error)
	GetAccounts() ([]*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	conn := "user=postgres dbname=postgres password=gobank sslmode=disable"
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	accounts := make([]*Account, 0, 10)
	query := `SELECT * from account`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		acc, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		// fmt.Println("Account", acc)
		accounts = append(accounts, acc)
	}
	// fmt.Println("Accounts", accounts)
	return accounts, nil
}

func (s *PostgresStore) CreateAccount(a *Account) error {
	resp, err := s.db.Query(`insert into account (first_name, last_name,number,balance,created_at) values ($1,$2,$3,$4,$5)`, a.FirstName, a.LastName, a.Number, a.Balance, a.CreatedAt)
	if err != nil {
		return err
	}
	fmt.Println("Account created", resp)
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	_, err := s.db.Exec("Delete from account where id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateAccount(*Account) error {
	return nil
}

func (s *PostgresStore) GetAccountById(id int) (*Account, error) {
	query := `select * from account where id = $1`
	res := s.db.QueryRow(query, id)
	acc := new(Account)

	err := res.Scan(&acc.ID, &acc.FirstName, &acc.LastName, &acc.Number, &acc.Balance, &acc.CreatedAt)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStore) GetAccountByJWT(string) (*Account, error) {
	return nil, nil
}

func (s *PostgresStore) createAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS account (
	id serial PRIMARY KEY,
	first_name varchar(50),
	last_name varchar(50),
	number serial,
	balance serial,
	created_at timestamp
)`

	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(&account.ID, &account.FirstName, &account.LastName, &account.Number, &account.Balance, &account.CreatedAt)

	return account, err
}
