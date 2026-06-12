package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Erro ao carregar o arquivo .env")
	}

	db := Connection()

	if err := Migrate(db); err != nil {
		panic(err)
	}

	var totalCeps int64
	if err := db.Model(&Cep{}).Count(&totalCeps).Error; err != nil {
		panic(err)
	}

	if totalCeps == 0 {
		log.Println("banco vazio, importando CEPs do arquivo cep.json")
		if err := PreloadCepsReadFileJSON(db, "cep.json"); err != nil {
			panic(err)
		}
	} else {
		log.Printf("banco já possui %d CEPs, pulando importação", totalCeps)
	}
	//

	httpServer := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cep := r.URL.Query().Get("cep")
			if cep == "" {
				http.Error(w, "cep is required", http.StatusBadRequest)
				return
			}

			var cepResult Cep
			if err := db.First(&cepResult, "cep = ?", cep).Error; err != nil {
				fmt.Println("error", err)
				http.Error(w, "cep not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			// retornar todo o resultado como JSON
			json, err := json.Marshal(cepResult)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			w.Write(json)
		}),
	}

	if err := httpServer.ListenAndServe(); err != nil {
		panic(err)
	}

	fmt.Println("Servidor iniciado na porta 8080")
}
