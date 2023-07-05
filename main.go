package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dotenv-org/godotenvvault"
	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
)

type Student struct {
	ID int
	Name string `validate:"required,min=3,max=50"`
	Age int64 `validate:"required,gte=0,lte=130"`
}


func dbConn()(db *sql.DB){
	driver := os.Getenv("DRIVER")
	user :=	os.Getenv("MYSQL_USER")
	password :=	os.Getenv("MYSQL_PASSWORD")
	host := os.Getenv("MYSQL_HOST")
	port :=	os.Getenv("PORT")
	dbname := os.Getenv("MYSQL_DATABASE")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user,password,host,port,dbname)
	db, err := sql.Open(driver, dsn)
	if err != nil {
		panic(err.Error())
	}
	return db
}


func main(){
	loadEnv()
	fmt.Println("========== GO MYSQL ==========")
	
	http.HandleFunc("/", Index)
	http.HandleFunc("/new", New)
	http.HandleFunc("/store", Store)
	http.HandleFunc("/show", Show)
	
	http.ListenAndServe(":8085", nil)
	
}


func Index(w http.ResponseWriter, r *http.Request){

	db := dbConn()

	selDb, err := db.Query("select * from student ORDER BY id desc")

	if err != nil {
		panic(err.Error())
	}

	student := Student{}
	res := []Student{}

	for selDb.Next() {
		var id int 
		var name string
		var age int64

		err = selDb.Scan(&id,&name,&age)

		if err != nil {
			panic(err.Error())
		}

		student.ID = id
		student.Name = name
		student.Age = age

		res = append(res, student)
	}
	tmpl.ExecuteTemplate(w, "Index", res)
	defer db.Close()
}

func New(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w,"New",nil)
}

func Store(w http.ResponseWriter, r *http.Request){
	db := dbConn()

	if r.Method == "POST" {

		validate := validator.New()
		name := strings.TrimSpace(r.FormValue("name"))
		age := strings.TrimSpace(r.FormValue("age"))
		if age == "" {
			age = "0"
		}

		studentNew := &Student{
			Name: name,
			Age: transformInt(age),
		}
		errValid := validate.Struct(studentNew)

		if errValid != nil {
			errors := make(map[string]string)
			for _, err := range errValid.(validator.ValidationErrors) {
				fmt.Println(getErrorMessage(err))
				field := err.StructField()
				errors[field] = getErrorMessage(err)
			}

			data := struct {
				Student Student
				Errors map[string]string
			}{
				Student: *studentNew,
				Errors: errors,
			}
			tmpl.ExecuteTemplate(w, "New",data)
			return
		}
		storeForm , err := db.Prepare("insert into student(name,age) values (?,?)")

		if err != nil {
			panic(err.Error())
		}
		storeForm.Exec(name,age)

		defer db.Close()
		http.Redirect(w, r,"/", 301)
	}
}

func Show(w http.ResponseWriter, r *http.Request) {
	db := dbConn()

	studentID := r.URL.Query().Get("id")

	dbShow, err := db.Query("select * from student where id=?", studentID)

	if err != nil {
		panic(err.Error())
	}

	student := Student{}
	for dbShow.Next() {
		var id int 
		var name string
		var age int64

		err = dbShow.Scan(&id,&name,&age)

		if err != nil {
			panic(err.Error())
		}

		student.ID = id
		student.Name = name
		student.Age = age
	}
	tmpl.ExecuteTemplate(w,"Show", student)
	defer dbShow.Close()

}

func loadEnv(){
	err := godotenvvault.Load(".env")
	if err != nil {
    log.Fatal("Error loading .env file")
  }
	if err != nil {
    log.Fatal("Error loading .env file")
  }
}

func transformInt(value string) int64{

	parsed, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		panic(err.Error())
	}

	return parsed
}

func getErrorMessage(err validator.FieldError) string {
	var field string
	switch err.Tag(){
		case "required" :
			if err.Field() == "Name" {
				field ="Nome"
			}else {
				field = "Idade"
			}
			return fmt.Sprintf("%s é obrigatório ", field)
		case "min":
			if err.Field() == "Name" {
				field ="Nome"
			}else {
				field = "Idade"
			}
			return fmt.Sprintf("%s dever ter no minimo %s caracteres ", field,err.Param())
		case "max": 
			if err.Field() == "Name" {
				field ="Nome"
			}else {
				field = "Idade"
			}
			return fmt.Sprintf("%s dever ter no maximo %s caracteres ", field,err.Param())
		default:
			return err.Error()	
	}
}


var tmpl = template.Must(template.ParseGlob("tmpl/*"))