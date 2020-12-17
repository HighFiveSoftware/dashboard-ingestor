package main

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"log"
	"os"
)



type Place struct {
	Id            uint32  `csv:"UID"`
	IsoCode       string  `csv:"iso2"`
	FipsId        string  `csv:"FIPS"`
	Admin         string  `csv:"Admin2"`
	ProvinceState string  `csv:"Province_State"`
	CountryRegion string  `csv:"Country_Region"`
	Latitude      float64 `csv:"Lat"`
	Longitude     float64 `csv:"Long_"`
	CombinedKey   string  `csv:"Combined_Key"`
	Population    uint32  `csv:"Population"`
}


func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file %v", err)
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error connecting to database %v", err)
	}
	defer conn.Close(context.Background())


	gh := githubFetcher{
		client: github.NewClient(nil),
		owner:  "CSSEGISandData",
		repo:   "COVID-19",
		branch: "master",
	}

	ingestor := ingestorService{
		conn: conn,
		gh:   &gh,
	}

	err = ingestor.ingestPlaces()
	if err != nil {
		log.Fatalf("Error ingesting places %v", err)
	}

	//err = ingestor.gh.getNumbers(CONFIRMED, US)
	//if err != nil {
	//	log.Fatalf("Error getting numbers %v", err)
	//}


	// rows, err := conn.Query(context.Background(), "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'")
	//if err != nil {
	//	log.Fatalf("Query failed %v", err)
	//}
	//for rows.Next() {
	//	var name string
	//	err := rows.Scan(&name)
	//	if err != nil {
	//		log.Fatalf("Error scanning row %v", err)
	//	}
	//	log.Printf("Table name: %s", name)
	//}
}
