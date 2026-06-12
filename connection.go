package main

import (
	"os"

	"gorm.io/driver/mysql"
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

func Connection() *gorm.DB {
	connectionString := os.Getenv("DATABASE_URL")

	if connectionString == "" {
		panic("Erro ao carregar variaveis de ambiente")
	}

	db, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		panic("erro ao conectar no banco de dados")
	}

	return db
}
