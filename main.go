package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const DBUrl = "postgres://avito_user:avito_password@localhost:5434/avito_db"

var db *pgx.Conn

type Segment struct {
	ID      int    `json:"id"`
	Segment string `json:"slug"`
}

func CreateSegment(c *gin.Context) {
	var segment Segment
	c.BindJSON(&segment)

	_, err := db.Exec(context.Background(), "INSERT INTO segments (segment_name) VALUES ($1)", segment.Segment)
	if err != nil {
		panic(err)
	}

	c.JSON(201, gin.H{"message": "Segment created successfully!"})
}

func DeleteSegment(c *gin.Context) {
	slug := c.Param("slug")
	_, err := db.Exec(context.Background(), "DELETE FROM segments WHERE segment_name=$1", slug)
	if err != nil {
		panic(err)
	}

	c.JSON(200, gin.H{"message": "Segment deleted successfully!"})
}

func getSegmentIDByUserID(UserID string) ([]int, error) {
	var segmentId []int

	query := "SELECT segment_id FROM user_segments WHERE user_id = $1;"
	intUserID, _ := strconv.Atoi(UserID)
	//result, err := db.Exec(context.Background(), `Delete from t where col=$1`, "val")
	rows, err := db.Query(context.Background(), query, intUserID)
	if err != nil {
		log.Fatal("Error querying database:", err)
	}
	log.Println(&rows)

	defer rows.Close()

	for rows.Next() {
		var segment int
		err := rows.Scan(&segment)
		log.Println(segment)
		if err != nil {
			log.Fatal("Error scanning rows:", err)
		}
		segmentId = append(segmentId, segment)
	}
	log.Println(segmentId)
	if err := rows.Err(); err != nil {
		log.Fatal("Error iterating rows:", err)
	}

	return segmentId, err

}

func AddUserToSegment(c *gin.Context) {

	var request struct {
		UserID         int      `json:"user_id"`
		AddSegments    []string `json:"add_segments"`
		RemoveSegments []string `json:"remove_segments"`
	}

	c.BindJSON(&request)

	if len(request.AddSegments) == 0 && len(request.RemoveSegments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No actions with segments"})
		return
	}

	if len(request.AddSegments) != 0 {
		addQuery := "SELECT s.id FROM segments s left join user_segments us on s.id = us.segment_id and user_id = $1 where us.segment_id is null and s.segment_name = any($2);"
		rows, err := db.Query(context.Background(), addQuery, request.UserID, request.AddSegments)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
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

			result += fmt.Sprintf("(%d, %d), ", request.UserID, answer)

		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		result = strings.TrimSuffix(result, ", ")
		log.Println(result)

		hueusQuery := fmt.Sprintf("insert into user_segments (user_id, segment_id) values %s;", result)

		_, err = db.Exec(context.Background(), hueusQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	if len(request.RemoveSegments) != 0 {
		getDeleteSegment := "SELECT s.id FROM segments s where s.segment_name = any($1);"
		deleteRows, err := db.Query(context.Background(), getDeleteSegment, request.RemoveSegments)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
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

		deleteQuery := "DELETE FROM user_segments WHERE segment_id = any($1) and user_id = $2;"

		_, err = db.Exec(context.Background(), deleteQuery, idSlice, request.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ok!"})
}

func GetActiveSegments(c *gin.Context) {
	userID := c.Param("id")

	//query := "select segment_id from user_segments where user_id = $1;"
	query := `SELECT s.segment_name
			  FROM segments s
			  INNER JOIN user_segments us ON s.id = us.segment_id
			  WHERE us.user_id = $1;`

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

	conn, err := pgx.Connect(context.Background(), DBUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	//defer conn.Close(context.Background())
	db = conn
	log.Println("Connected successfully!")

}

func main() {
	NewClient()
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Привет, я Gin",
		})
	})

	r.POST("/segment", CreateSegment)
	r.POST("/user2segment", AddUserToSegment)
	r.GET("/segments/:id", GetActiveSegments)
	r.DELETE("/segment/:slug", DeleteSegment)
	r.Run(":8000")

}
