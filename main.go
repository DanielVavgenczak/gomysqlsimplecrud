package main

import (
	"database/sql"
	"fmt"
	handleerrors "gomysqlsimplecrud/handleErrors"
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

type PageData struct {
	Message string
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
	http.HandleFunc("/edit", Edit)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/delete", Delete)
	http.HandleFunc("/destroy", Destroy)

	http.ListenAndServe(":8085", nil)
	
}


func Index(w http.ResponseWriter, r *http.Request){

	db := dbConn()

	selDb, err := db.Query("select * from student ORDER BY id desc")

	if err != nil {
		panic(err.Error())
	}

	student := Student{}
	data := struct{
		Student []Student
		Counted int
	}{}

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

		data.Student = append(data.Student, student)
	}
	data.Counted = len(data.Student)
	
	tmpl.ExecuteTemplate(w, "Index", data)
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


func Edit(w http.ResponseWriter, r *http.Request){

	dbEdit := dbConn()

	data := struct {
		Student Student
		Errors map[string]string
	}{}

	eId := r.URL.Query().Get("id")
	selectEdit, err := dbEdit.Query("select * from student where id=?", eId)

	if err != nil{
		panic(err.Error())
	}
	
	for selectEdit.Next() {
		var id int
		var name string
		var age int64

		err = selectEdit.Scan(&id,&name,&age)
		if err != nil {
			panic(err.Error())
		}
		data.Student.ID = id
		data.Student.Name = name
		data.Student.Age = age
	}

	if data.Student.ID == 0 {
		data.Errors = map[string]string{"message":"Nenhum estudante encontrado!"}
		fmt.Print("Aqui",data.Student.ID, data.Errors)
		tmpl.ExecuteTemplate(w,"Edit", data)
		return
	}
		
	tmpl.ExecuteTemplate(w,"Edit", data)
	defer selectEdit.Close()
}

func Update(w http.ResponseWriter, r *http.Request) {
	dbEdit := dbConn()
	if r.Method == "POST" {
		name := strings.TrimSpace(r.FormValue("name"))
		age := strings.TrimSpace(r.FormValue("age"))
		id := strings.TrimSpace(r.FormValue("uid"))
		if age == "" {
			age = "0"
		}
		student := Student{
			Name: name,
			Age:  transformInt(age),
	}
		errValid := handleerrors.ValidationInputs(student)
		errors := []string{}
		if errValid != nil {			
			errors = append(errors, errValid.Error())
			dataUpdate := struct{
				Errors []string
			}{
				Errors: errors,
			}
			tmpl.ExecuteTemplate(w, "Edit",dataUpdate)
			return
		}

		updateInfo, err := dbEdit.Prepare("update student set name=?,age=? where id=?")
		
		if err != nil {
			panic(err.Error())
		}
		updateInfo.Exec(name, age, id)
	}
	  // Armazenar dados em um cookie
    cookie := &http.Cookie{
			Name:  "successData",
			Value: "Dados de sucesso aqui",
	}

	defer dbEdit.Close()
	http.SetCookie(w,cookie)
	http.Redirect(w,r,"/",301)

}


func Delete(w http.ResponseWriter, r *http.Request){

	dbEdit := dbConn()

	data := struct {
		Student Student
		Errors map[string]string
	}{}

	dId := r.URL.Query().Get("id")
	deleteStudent, err := dbEdit.Query("select * from student where id=?", dId)

	if err != nil{
		panic(err.Error())
	}
	
	for deleteStudent.Next() {
		var id int
		var name string
		var age int64

		err = deleteStudent.Scan(&id,&name,&age)
		if err != nil {
			panic(err.Error())
		}
		data.Student.ID = id
		data.Student.Name = name
		data.Student.Age = age
	}

	if data.Student.ID == 0 {
		data.Errors = map[string]string{"message":"Nenhum estudante encontrado!"}
		fmt.Print("Aqui",data.Student.ID, data.Errors)
		tmpl.ExecuteTemplate(w,"Edit", data)
		return
	}
		
	tmpl.ExecuteTemplate(w,"Delete", data)
	defer deleteStudent.Close()
}

func Destroy(w http.ResponseWriter, r *http.Request) {

	db := dbConn()
    emp := r.URL.Query().Get("id")
    delForm, err := db.Prepare("delete from student where id=?")
    if err != nil {
        panic(err.Error())
    }
    delForm.Exec(emp)
    defer db.Close()
    http.Redirect(w, r, "/", 301)
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