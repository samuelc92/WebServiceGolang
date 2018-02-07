package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Conta struct {
	Cnpj []Php `json:"cnpj"`
}

type Php struct {
	Cnpj string `json:"Cnpj"`
}

type ContaRetorno struct {
	Cnpj []RetornoPhp `json:"cnpj"`
}
type RetornoPhp struct {
	Cnpj     string `json:"Cnpj"`
	DataVenc string `json:"dataVenc"`
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	log.Fatal(http.ListenAndServe(":9091", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	retorno, err := Listar()

	if err != "" {
		fmt.Fprintf(w, "Fail: %q", err)
	} else {
		for _, v := range retorno.Cnpj {
			location, _ := time.LoadLocation("Local")
			var data time.Time
			if v.DataVenc == "" {
				data = time.Date(1, time.January, 1, 0, 0, 0, 0, location)
			} else {
				fmt.Println(v.DataVenc)
				dataAux := strings.Split(v.DataVenc, "/")
				year, _ := strconv.Atoi(dataAux[2])
				//month, _ := strconv.Atoi(dataAux[1])
				day, _ := strconv.Atoi(dataAux[0])
				data = time.Date(year, time.November, day, 0, 0, 0, 0, location)
			}
			fmt.Fprintf(w, "CNPJ: %q - dataVenc: %q  ", v.Cnpj, data)
		}
	}
}

func Listar() (*ContaRetorno, string) {
	db, err := sql.Open("mysql", "root:root@tcp(192.168.1.69:3306)/lifesysconta")
	if err != nil {
		return nil, "Fail"
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, "Fail ping" // proper error handling instead of panic in your app
	}

	rows, err := db.Query("SELECT Cnpj FROM conta")

	if err != nil {
		return nil, "Fail cnpj" // proper error handling instead of panic in your app
	}

	defer rows.Close()
	//arrCnpj := make([]string, 0)
	dados := &Conta{Cnpj: make([]Php, 0)}
	for rows.Next() {
		var cnpj string
		err := rows.Scan(&cnpj)
		if err != nil {
			return nil, "Fail read CNPJ" // proper error handling instead of panic in your app
		}
		dados.Cnpj = append(dados.Cnpj, Php{Cnpj: cnpj})
	}

	sistemaAtendimento := "http://localhost.com.br/"
	jsonValue, err := json.Marshal(dados)
	if err != nil {
		fmt.Println("Error Marshal")
	}
	req, err := http.NewRequest(http.MethodPost, sistemaAtendimento,
		bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Println("Error Request")
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("Error Response")
	}
	fmt.Println(resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "Fail readall retornoPhp"
	}
	
	var retornoPhp ContaRetorno
	errorr := json.Unmarshal(body, &retornoPhp)

	if errorr != nil {
		return nil, "Fail decode retornophp"
	}

	defer resp.Body.Close()

	return &retornoPhp, ""
}
