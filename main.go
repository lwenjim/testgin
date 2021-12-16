package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
	"github.com/tealeg/xlsx/v3"
	"github.com/thedevsaddam/gojsonq/v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Response struct {
	Code int                   `json:"code"`
	Msg  string                `json:"msg"`
	Data []Course              `json:"data"`
	Meta map[string]Pagination `json:"meta"`
}

type User struct {
	UserId        int         `json:"user_id"`
	UserName      interface{} `json:"user_name"`
	Email         interface{} `json:"email"`
	RealName      interface{} `json:"real_name"`
	Salt          interface{} `json:"salt"`
	Sex           int         `json:"sex"`
	Mobile        interface{} `json:"mobile"`
	Birthday      interface{} `json:"birthday"`
	CreatedUser   int         `json:"created_user"`
	CreatedAt     interface{} `json:"created_at"`
	UpdatedUser   int         `json:"updated_user"`
	UpdatedAt     interface{} `json:"updated_at"`
	LastLoginIp   interface{} `json:"last_login_ip"`
	LastLoginTime interface{} `json:"last_login_time"`
	Status        int         `json:"status"`
	Autoutime     interface{} `json:"autoutime"`
}

type Course struct {
	Id           int         `json:"id"`
	CourseName   interface{} `json:"course_name"`
	SubjectId    int         `json:"subject_id"`
	Type         int         `json:"type"`
	Season       interface{} `json:"season"`
	Stage        int         `json:"stage"`
	Grade        int         `json:"grade"`
	Textbook     int         `json:"textbook"`
	Usage        int         `json:"usage"`
	Description  interface{} `json:"description"`
	Scene        int         `json:"scene"`
	Status       int         `json:"status"`
	CreatedTime  interface{} `json:"created_time"`
	CreateUser   int         `json:"create_user"`
	UpdatedTime  interface{} `json:"updated_time"`
	UpdateUser   int         `json:"update_user"`
	Autoutime    interface{} `json:"autoutime"`
	ChannelId    int         `json:"channel_id"`
	KnowledgeNum int         `json:"knowledge_num"`
	SectionId    int         `json:"section_id"`

	MapModuleId []int  `json:"map_module_id"`
	DeletedTime string `json:"deleted_time"`
	CreatedUser User   `json:"created_user"`
	UpdatedUser User   `json:"updated_user"`
}

type Pagination struct {
	Total       int               `json:"total"`
	Count       int               `json:"count"`
	PerPage     int               `json:"per_page"`
	CurrentPage int               `json:"current_page"`
	TotalPages  int               `json:"total_pages"`
	Links       map[string]string `json:"links"`
}

type LocalCourse struct {
	CourseId   int64  `json:"course_id"`
	CourseName string `json:"course_name"`
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := "Hello World!"
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	log.Printf(" [x] Sent %s", body)
	failOnError(err, "Failed to publish a message")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
func localSheet() {
	jq := gojsonq.New().File("./device_list.txt").From("dList").Select("deviceName", "deviceId", "position")
	deviceInfoList, ok := jq.Get().([]interface{})
	if !ok {
		fmt.Println("Convert deviceInfoList error")
	}
	xlsxFile := xlsx.NewFile()
	sheet, err := xlsxFile.AddSheet("Sheet 1")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	sheet.AddRow().WriteSlice(&[]string{"设备名称", "设备ID", "位置"}, 3)
	for _, deviceInfo := range deviceInfoList {
		deviceInfoMap, ok := deviceInfo.(map[string]interface{})
		if !ok {
			fmt.Println("Convert deviceInfoMap error")
		}
		row := sheet.AddRow()
		row.AddCell().SetValue(deviceInfoMap["deviceName"])
		row.AddCell().SetValue(deviceInfoMap["deviceId"])
		row.AddCell().SetValue(deviceInfoMap["position"])
	}
	xlsxFile.Save("./result.xlsx")
}

func localMysql() {
	db, err := sql.Open("mysql", "lwenjim:111111@/course_db")
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	rows, err := db.Query("select course_id,course_name from course limit 100")
	if err != nil {
		panic(err)
	}
	var data []*LocalCourse
	for rows.Next() {
		_course := new(LocalCourse)
		rows.Scan(&_course.CourseId, &_course.CourseName)
		data = append(data, _course)
	}
	buffer, _ := json.Marshal(data)
	txt := string(buffer)
	buffer3 := gojsonq.New().FromString(txt).Select("course_id", "course_name").Get()
	buffer, _ = json.MarshalIndent(buffer3, "", " ")
	fmt.Println(string(buffer))
}

func localHttp() {
	request := http.Client{}
	response, err := request.Get("http://dev-course-v5.classba.cn/course/course")
	if err != nil {
		println(err)
		return
	}
	defer response.Body.Close()
	buffer, err := ioutil.ReadAll(response.Body)
	if err != nil {
		println(err)
		return
	}
	resp := Response{}
	if err := json.Unmarshal(buffer, &resp); err != nil {
		println(err)
		return
	}
	var _course Course
	for _, _course = range resp.Data {
		if newString, ok := _course.CourseName.(string); ok {
			println(fmt.Sprintf("%d, %s", _course.Id, newString))
		} else {
			println(fmt.Sprintf("%d, %s", _course.Id, strconv.Itoa((int)(_course.CourseName.(float64)))))
		}
	}
}

func localRedis() {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(context.Background(), cursor, "*", 20).Result()
		if err != nil {
			panic(err)
		}
		for _, key := range keys {
			fmt.Println(key)
		}
		if cursor == 0 {
			break
		}
	}
}
