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
	GetAccountByID(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "postgres://postgres:password@0.0.0.0/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.CreateAcccountTable()
}

func (s *PostgresStore) CreateAcccountTable() error {
	query := `create table account (
    			id serial primary key ,
                first_name varchar(50),
                     last_name varchar(50),
                     number serial,
                     balance int,
                     created_at timestamp
)`
	_, err := s.db.Exec(query)
	return err

}

func (s *PostgresStore) CreateAccount(account *Account) error {
	query := `insert into account(first_name,last_name,number,balance,created_at) values 
                        ($1,$2,$3,$4,$5)`
	resp, err := s.db.Query(query, account.FirstName, account.LastName, account.Number, account.Balance, account.CreatedAt)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	return nil
}

func (s *PostgresStore) UpdateAccount(account *Account) error {
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	return nil
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	return nil, nil
}
