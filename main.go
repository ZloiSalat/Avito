package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"os"
	"strings"
)

const DBUrl = "postgres://avito_user:avito_password@db:5432/avito_db"

type Segment struct {
	ID      int    `json:"id"`
	Segment string `json:"slug"`
}

var db *pgx.Conn

func CreateSegment(c *gin.Context) {

	var segment Segment
	err := c.BindJSON(&segment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to create a request"})
		return
	}
	//Запрос на внесение названия сегмента в столбец segment_name в таблице segments.
	_, err = db.Exec(context.Background(), "INSERT INTO segments (segment_name) VALUES ($1)", segment.Segment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to create a segment"})
		return
	}

	c.JSON(201, gin.H{"message": "Segment created successfully!"})
}

func DeleteSegment(c *gin.Context) {

	slug := c.Param("slug")
	//Запрос на удаление сегмента по названию (slug) сегмента.
	_, err := db.Exec(context.Background(), "delete from segments where segment_name=$1", slug)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to delete a request"})
		return
	}

	c.JSON(200, gin.H{"message": "Segment deleted successfully!"})
}

func AddUserToSegment(c *gin.Context) {

	var request struct {
		UserID         int      `json:"user_id"`
		AddSegments    []string `json:"add_segments"`
		RemoveSegments []string `json:"remove_segments"`
	}

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to create a request"})
		return
	}
	//Проверка на заполненность строк.
	if len(request.AddSegments) == 0 && len(request.RemoveSegments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No actions with segments"})
		return
	}

	if len(request.AddSegments) != 0 {
		//Запрос на проверку существующих сегментов, чтобы отсеять повторяющиеся сегменты.
		addQuery := "select s.id FROM segments s left join user_segments us on s.id = us.segment_id and user_id = $1 where us.segment_id is null and s.segment_name = any($2);"
		rows, err := db.Query(context.Background(), addQuery, request.UserID, request.AddSegments)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Query message error": err.Error()})
			return
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

		_, err = db.Exec(context.Background(), insertQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Insert message error": err.Error()})
			return
		}
	}

	if len(request.RemoveSegments) != 0 {
		//Запрос на проверку существующих сегментов, чтобы отсеять повторяющиеся сегменты.
		getDeleteSegment := "select s.id from segments s where s.segment_name = any($1);"
		deleteRows, err := db.Query(context.Background(), getDeleteSegment, request.RemoveSegments)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"getDelete message error: ": err.Error()})
			return
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

		_, err = db.Exec(context.Background(), deleteQuery, idSlice, request.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Delete message error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ok!"})
}

func GetActiveSegments(c *gin.Context) {
	userID := c.Param("id")

	query := `select s.segment_name
			  from segments s
			  inner join user_segments us on s.id = us.segment_id
			  where us.user_id = $1;`
	//Получение активных сегментов по номеру id
	rows, err := db.Query(context.Background(), query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var segments []string
	for rows.Next() {
		var segment string
		err := rows.Scan(&segment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		segments = append(segments, segment)
	}
	c.JSON(http.StatusOK, gin.H{"segments": segments})
}

func NewClient() {
	//Подключение к базе данных
	conn, err := pgx.Connect(context.Background(), DBUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	db = conn
	log.Println("Connected successfully!")

}

func main() {
	NewClient()
	r := gin.Default()

	r.POST("/segment", CreateSegment)
	r.POST("/manageUserSegments", AddUserToSegment)
	r.GET("/segments/:id", GetActiveSegments)
	r.DELETE("/segment/:slug", DeleteSegment)
	err := r.Run(":8000")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open a port: %v\n", err)
		return
	}

}
