package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Cep{}); err != nil {
		return err
	}
	return nil
}

func PreloadCepsReadFileJSON(db *gorm.DB, filePath string) error {
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
