package main

import (
    "fmt"
    "net/http"
    "log"
    "encoding/json"

    "github.com/julienschmidt/httprouter"
    _ "github.com/go-sql-driver/mysql"
    "github.com/jmoiron/sqlx"
)

type Person struct {
    Cedula string `db:"cedula"`
    Nombre  string `db:"nombre"`
}
type Error struct {
    Success bool `json:"success"`
    Error string `json:"error"`
}

func GetAlbums(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    w.Header().Set("Content-Type", "application/json")
    db, err := sqlx.Connect("mysql", "teti:Lautaro123.@(localhost:3306)/Agribussiness")
    if err != nil {
        log.Fatalln(err)
    }

    people := []Person{}
    err = db.Select(&people, "SELECT nombre, cedula FROM Personas ORDER BY cedula ASC")
    if err != nil {
        json.NewEncoder(w).Encode(Error{
            Success: false,
            Error: err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(people)
}

func main() {

    router := httprouter.New()
    router.GET("/albums/", GetAlbums)
    fmt.Println("Server started")

    log.Fatal(http.ListenAndServe(":8080", router))
}
