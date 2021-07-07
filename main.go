package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
    "fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type Post struct {
	id_usuario string `json:"id_usuario"`
	nombre     string `json:"nombre"`
	paterno    string `json:"paterno"`
	materno    string `json:"materno"`
	telefono   string `json:"telefono"`
	email      string `json:"email"`
	password   string `json:"password"`
}

type Movimientos struct {
	id_usuario    string `json:"id_usuario"`
	nombre        string `json:"nombre"`
	paterno       string `json:"paterno"`
	materno       string `json:"materno"`
	numero_cuenta string `json:"numero_cuenta"`
	saldo         string `json:"saldo"`
	movimiento    string `json:"movimiento"`
	importe       string `json:"importe"`
}

type Transferencia struct {
	id_usuario        string `json:"id_usuario"`
	importe           string `json:"importe"`
	no_cuenta_destino string `json:"no_cuenta_destino"`
}

type Datos struct {
	id_usuario        string `json:"id_usuario"`
	saldo             string `json:"saldo"`
	id_cuenta         string `json:"no_cuenta"`
}

type Deposito struct {
	id_usuario string `json:"id_usuario"`
	nombre    string `json:"nombre"`
	paterno    string `json:"paterno"`
	materno    string `json:"materno"`
	importe    string `json:"importe"`
	no_cuenta  string `json:"no_cuenta"`
}

var db *sql.DB
var err error

func main() {
	// db, err = sql.Open("mysql", "<ksolis>:<cnTsÑ1020.>@tcp(127.0.0.1:3306)/<pandita>")

	db, err = sql.Open("mysql", "ksolis:cnTsÑ1020.@tcp(127.0.0.1:3306)/pandita?charset=utf8&parseTime=True&loc=Local")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/posts", getMovimientos).Methods("GET")
	router.HandleFunc("/posts", createUsuarios).Methods("POST")
	router.HandleFunc("/posts/{id_usuario}", getMovUser).Methods("GET")
	router.HandleFunc("/posts/{id_usuario}", updateUsuario).Methods("PUT")
	router.HandleFunc("/posts/{id_usuario}", deleteUsuario).Methods("DELETE")
	http.ListenAndServe(":8000", router)
}

func getMovimientos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var posts []Movimientos
	result, err := db.Query("SELECT id_usuario, nombre, paterno, materno, cuentas.no_cuenta, cuentas.saldo, movimientos.movimiento, movimientos.importe FROM usuarios inner join cuentas ON usuarios.id_usuario = cuentas.fk_id_usuario inner join movimientos ON cuentas.id_cuenta = movimientos.fk_id_cuenta")

	if err != nil {
		panic(err.Error())
	}

	defer result.Close()

	for result.Next() {
		var mov Movimientos
		err := result.Scan(&mov.id_usuario, &mov.nombre, &mov.paterno, &mov.materno, &mov.numero_cuenta, &mov.saldo, &mov.movimiento, &mov.importe)

		if err != nil {
			panic(err.Error())
		}
		posts = append(posts, mov)

	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(posts)
}

func createUsuarios(w http.ResponseWriter, r *http.Request) {

	// stmt, err := db.Prepare("INSERT INTO usuarios(nombre, paterno, materno, telefono, email, fecha_registro, hash, activo) VALUES(?,?,?,?,?,?,?,?)")

	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}
	// cuentaRand := createRand()
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	nombre := keyVal["nombre"]
	paterno := keyVal["paterno"]
	materno := keyVal["materno"]
	telefono := keyVal["telefono"]
	email := keyVal["email"]
	password := keyVal["password"]
	hash, _ := HashPassword(password)
	fecha_registro := time.Now()
	activo := 1

	x, err := db.Exec("INSERT INTO usuarios(nombre, paterno, materno, telefono, email, fecha_registro, hash, activo) VALUES(?,?,?,?,?,?,?,?)", nombre, paterno, materno, telefono, email, fecha_registro, hash, activo)

	if err != nil {
		panic(err.Error())
	}
	// cuenta = int(x.LastInsertId())
	var id int64
	id, err = x.LastInsertId()
	if err != nil {
		fmt.Println(err)
	}

	insertado := int(id)
	fmt.Println(insertado)
	x, err = db.Exec("INSERT INTO cuentas(no_cuenta,saldo,fk_id_usuario, fecha_registro) VALUES(?,?,?,?)", createRand(), 1000, insertado, fecha_registro)
	//x.LastInsertId() createRand()
	if err != nil {
		fmt.Println(err)
	}

	var id_mov int64
	id_mov, err = x.LastInsertId()
	if err != nil {
		fmt.Println(err)
	}

	insertadoMov := int(id_mov)
	fmt.Println(insertadoMov)
	db.Exec("INSERT INTO movimientos (movimiento,importe,fk_id_cuenta,fecha_registro,activo, tipo) VALUES(?,?,?,?,?)", "Regalo", 1000, insertadoMov, fecha_registro, 1, 1)

	fmt.Fprintf(w, "New post was created")
}

func getMovUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	result, err := db.Query("SELECT id_usuario, nombre FROM usuarios inner join cuentas ON usuarios.id_usuario = cuentas.fk_id_usuario inner join movimientos ON cuentas.id_cuenta = movimientos.fk_id_movimiento  WHERE id_usuario = ?", params["id_usuario"])
	if err != nil {
		panic(err.Error())
	}
	defer result.Close()
	var post Post
	for result.Next() {
		err := result.Scan(&post.id_usuario, &post.nombre)
		if err != nil {
			panic(err.Error())
		}
	}
	json.NewEncoder(w).Encode(post)
}

func updateUsuario(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	stmt, err := db.Prepare("UPDATE usuarios SET nombre = ?, paterno = ?, materno = ? , telefono = ? WHERE id_usuario = ?")
	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	newNombre := keyVal["nombre"]
	newPaterno := keyVal["paterno"]
	newMaterno := keyVal["materno"]
	newTelefono := keyVal["telefono"]
	_, err = stmt.Exec(newNombre, newPaterno, newMaterno, newTelefono, params["id_usuario"])
	if err != nil {
		panic(err.Error())
	}
	fmt.Fprintf(w, "Post with id_usuario = %s was updated", params["id_usuario"])
}

func deleteUsuario(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	stmt, err := db.Prepare("DELETE FROM usuarios WHERE id_usuario = ?")
	if err != nil {
		panic(err.Error())
	}
	_, err = stmt.Exec(params["id_usuario"])
	if err != nil {
		panic(err.Error())
	}
	fmt.Fprintf(w, "Post with id_usuario = %s was deleted", params["id_usuario"])
}



func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func createRand() int {
	rand.Seed(time.Now().UnixNano())
	cuenta := 0

	cuenta += randomInt(10000000, 99999999)
	return cuenta
}


