package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	CreateCommand(*Command) error
	DeleteCommand(int) error
	GetCommands() ([]*Command, error)
	CreateWorker(*Worker) error
	GetWorkers() ([]*Worker, error)
	Register(string, string) (*Worker, error)
	GetWorkerByEmail(email string) (*Worker, error)
	GetAccountByID(id string) (*Worker, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "host=postgresql://postgreskrixo_user:bFaObNSTztS3ZlgJ8uGt7F2SF7gfH8Uz@dpg-d1fjthvfte5s73fk4sgg-a.frankfurt-postgres.render.com/postgreskrixo port=5432 user=postgreskrixo_user dbname=postgreskrixo password=bFaObNSTztS3ZlgJ8uGt7F2SF7gfH8Uz sslmode=disable"

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
	if err := s.createCommandTable(); err != nil {
		return err
	}

	if err := s.createWorkersTable(); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) createCommandTable() error {
	query := `create table if not exists command (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		fullname varchar(100),
		number serial,
		services varchar(100),
		workers serial,
		start varchar(100),
		distination varchar(100)
	);`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateCommand(acc *Command) error {
	query := `insert into command 
	(fullname, number, services, workers, start, distination)
	values ($1, $2, $3, $4, $5, $6)`

	_, err := s.db.Query(
		query,
		acc.FullName,
		acc.Number,
		acc.Service,
		acc.Workers,
		acc.Start,
		acc.Distination)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) DeleteCommand(id int) error {
	_, err := s.db.Query("delete from command where id = $1", id)
	return err
}

func (s *PostgresStore) GetCommands() ([]*Command, error) {
	rows, err := s.db.Query("select * from command ")
	if err != nil {
		return nil, err
	}

	accounts := []*Command{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}
func (s *PostgresStore) createWorkersTable() error {
	query := `CREATE TABLE IF NOT EXISTS worker (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fullname VARCHAR(100) NOT NULL ,
    number VARCHAR(20) NOT NULL ,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL ,
    position VARCHAR(100) NOT NULL ,
    experience TEXT NOT NULL ,
    message TEXT NOT NULL ,
    isaccepted BOOLEAN NOT NULL
);`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateWorker(worker *Worker) error {
	hashedpassword, err := s.CreateUser(worker.Email, worker.Password)
	if err != nil {
		return err
	} else {
		query := `INSERT INTO worker (fullname, number, email, password, position, experience, message, isaccepted) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8); `

		_, err := s.db.Query(query, worker.FullName, worker.Number, worker.Email, hashedpassword, worker.Position, worker.Experience, worker.Message, worker.IsAccepted)

		return err
	}
}

func (s *PostgresStore) GetWorkers() ([]*Worker, error) {
	rows, err := s.db.Query("select * from worker ")
	if err != nil {
		return nil, err
	}

	Workers := []*Worker{}
	for rows.Next() {
		worker, err := scanIntoWorker(rows)
		if err != nil {
			return nil, err
		}
		Workers = append(Workers, worker)
	}

	return Workers, nil
}

// func (s *PostgresStore) Register(password string, email string) (bool, error) {
// 	var hashedPassword string

// 	// Fetch the hashed password from the database
// 	query := `SELECT password FROM worker WHERE email = $1`
// 	err := s.db.QueryRow(query, email).Scan(&hashedPassword)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			// User not found
// 			return false, nil
// 		}
// 		// Other DB error
// 		return false, err
// 	}

// 	// Compare the plaintext password with the stored hash
// 	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
// 	if err != nil {
// 		// Password mismatch
// 		return false, nil
// 	}

// 	// Password matches
// 	return true, nil
// }

func (s *PostgresStore) Register(password string, email string) (*Worker, error) {
	worker, err := s.GetWorkerByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(worker.Password), []byte(password))
	if err != nil {
		// Password mismatch
		return nil, nil
	}

	return worker, nil
}

func (s *PostgresStore) GetWorkerByEmail(email string) (*Worker, error) {
	rows, err := s.db.Query("select * from worker where email = $1", email)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoWorker(rows)
	}

	return nil, fmt.Errorf("account %s not found", email)
}

func (s *PostgresStore) GetAccountByID(id string) (*Worker, error) {
	rows, err := s.db.Query("select * from worker where id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoWorker(rows)
	}

	return nil, fmt.Errorf("account %s not found", id)
}

func scanIntoAccount(rows *sql.Rows) (*Command, error) {
	command := new(Command)
	err := rows.Scan(
		&command.ID,
		&command.FullName,
		&command.Number,
		&command.Service,
		&command.Workers,
		&command.Start,
		&command.Distination)

	return command, err
}

func scanIntoWorker(rows *sql.Rows) (*Worker, error) {
	worker := new(Worker)
	err := rows.Scan(
		&worker.ID,
		&worker.FullName,
		&worker.Number,
		&worker.Email,
		&worker.Password,
		&worker.Position,
		&worker.Experience,
		&worker.Message,
		&worker.IsAccepted)

	return worker, err
}
