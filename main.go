package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Cep struct {
	gorm.Model
	Cep          string `json:"cep" gorm:"index"`
	Logradouro   string `json:"logradouro"`
	Localidade   string `json:"localidade"`
	CodMunicipio string `json:"id_municipio"`
	Municipio    string `json:"nome_municipio"`
	Uf           string `json:"sigla_uf"`
	Point        string `json:"centroide"`
}

func migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Cep{}); err != nil {
		return err
	}
	return nil
}

func preloadCepsReadFileJSON(db *gorm.DB, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Aumenta o limite caso alguma linha JSON seja grande.
	scanner.Buffer(make([]byte, 1024), 10*1024*1024)

	const batchSize = 1000

	ceps := make([]Cep, 0, batchSize)
	linha := 0
	total := 0

	for scanner.Scan() {
		linha++

		texto := scanner.Text()
		if texto == "" {
			continue
		}

		var cep Cep
		if err := json.Unmarshal([]byte(texto), &cep); err != nil {
			log.Printf("erro ao ler JSON na linha %d: %v", linha, err)
			continue
		}

		ceps = append(ceps, cep)

		if len(ceps) >= batchSize {
			if err := db.CreateInBatches(ceps, batchSize).Error; err != nil {
				return err
			}

			total += len(ceps)
			log.Printf("%d CEPs importados", total)
			ceps = ceps[:0]
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(ceps) > 0 {
		if err := db.CreateInBatches(ceps, batchSize).Error; err != nil {
			return err
		}

		total += len(ceps)
		log.Printf("%d CEPs importados", total)
	}

	return nil
}

func main() {
	db, err := gorm.Open(sqlite.Open("cepdatabase.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := migrate(db); err != nil {
		panic(err)
	}

	var totalCeps int64
	if err := db.Model(&Cep{}).Count(&totalCeps).Error; err != nil {
		panic(err)
	}

	if totalCeps == 0 {
		log.Println("banco vazio, importando CEPs do arquivo cep.json")
		if err := preloadCepsReadFileJSON(db, "cep.json"); err != nil {
			panic(err)
		}
	} else {
		log.Printf("banco já possui %d CEPs, pulando importação", totalCeps)
	}

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

}
