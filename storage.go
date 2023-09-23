package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"strings"
)

type Storage interface {
	CreateSegment(*User) error
	DeleteSegment(string) error
	AddUserToSegment(*Request) error
	GetActiveSegments(int) (*User, error)
}

type PostgresStore struct {
	db *pgx.Conn
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "postgres://avito_user:avito_password@localhost:5434/avito_db"
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	//defer conn.Close(context.Background())
	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}
	return &PostgresStore{
		db: conn,
	}, nil
}

func (s *PostgresStore) CreateSegment(u *User) error {
	query := "insert into segments (segment_name) values ($1)"
	resp, err := s.db.Exec(context.Background(), query, u.Segment)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", resp)
	return nil
}

func (s *PostgresStore) DeleteSegment(slug string) error {
	query := "delete from segments where segment_name=$1"
	_, err := s.db.Query(context.Background(), query, slug)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) AddUserToSegment(request *Request) error {

	//Проверка на заполненность строк.
	if len(request.AddSegments) == 0 && len(request.RemoveSegments) == 0 {
		return fmt.Errorf("fail to read data")
	}

	if len(request.AddSegments) != 0 {
		//Запрос на проверку существующих сегментов, чтобы отсеять повторяющиеся сегменты.
		addQuery := "select s.id FROM segments s left join user_segments us on s.id = us.segment_id and user_id = $1 where us.segment_id is null and s.segment_name = any($2);"
		rows, err := s.db.Query(context.Background(), addQuery, request.UserID, request.AddSegments)
		if err != nil {
			return err
		}

		defer rows.Close()

		var result string
		for rows.Next() {
			var answer int
			err := rows.Scan(&answer)
			if err != nil {
				log.Fatal(err)
			}
			//Форматирование результатов
			result += fmt.Sprintf("(%d, %d), ", request.UserID, answer)

		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		result = strings.TrimSuffix(result, ", ")
		//Запрос на внесение данных по значениям user_id и segment_id
		insertQuery := fmt.Sprintf("insert into user_segments (user_id, segment_id) values %s;", result)

		_, err = s.db.Exec(context.Background(), insertQuery)
		if err != nil {
			return err
		}
	}

	if len(request.RemoveSegments) != 0 {
		//Запрос на проверку существующих сегментов, чтобы отсеять повторяющиеся сегменты.
		getDeleteSegment := "select s.id from segments s where s.segment_name = any($1);"
		deleteRows, err := s.db.Query(context.Background(), getDeleteSegment, request.RemoveSegments)
		if err != nil {
			return err
		}

		defer deleteRows.Close()

		var idSlice []int

		for deleteRows.Next() {
			var deleteAnswer int
			err := deleteRows.Scan(&deleteAnswer)
			if err != nil {
				log.Fatal(err)
			}

			idSlice = append(idSlice, deleteAnswer)
		}
		if err := deleteRows.Err(); err != nil {
			log.Fatal(err)
		}
		//Удаление сегментов
		deleteQuery := "delete from user_segments where segment_id = any($1) and user_id = $2;"

		_, err = s.db.Exec(context.Background(), deleteQuery, idSlice, request.UserID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStore) GetActiveSegments(id int) (*User, error) {
	query := `select s.segment_name
			  from segments s
			  inner join user_segments us on s.id = us.segment_id
			  where us.user_id = $1;`
	rows, err := s.db.Query(context.Background(), query, id)
	if err != nil {
		return nil, err
	}

	var concatenatedSegment string // Строка для конкатенации сегментов.

	for rows.Next() {
		var segmentName string
		if err := rows.Scan(&segmentName); err != nil {
			return nil, err
		}
		concatenatedSegment += segmentName + "," // Конкатенируем сегменты с запятой в качестве разделителя.
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Удалите последнюю запятую, если она есть.
	if len(concatenatedSegment) > 0 && concatenatedSegment[len(concatenatedSegment)-1] == ',' {
		concatenatedSegment = concatenatedSegment[:len(concatenatedSegment)-1]
	}

	return &User{
		Segment: concatenatedSegment,
	}, nil

}
