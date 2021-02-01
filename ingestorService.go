package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"strconv"
	"time"
)

type ingestorService struct {
	conn *pgx.Conn
	gh   *githubFetcher
}

func (ingestor *ingestorService) ingestPlaces() error {
	places, err := ingestor.gh.getPlaces()
	if err != nil {
		return err
	}

	_, err = ingestor.conn.CopyFrom(
		context.Background(),
		pgx.Identifier{"places"},
		[]string{"id", "iso_code", "fips_id", "admin", "province_state", "country_region", "coordinate", "combined_key", "population"},
		pgx.CopyFromSlice(len(places), func(i int) ([]interface{}, error) {
			p := places[i]
			var loc pgtype.Point
			if p.Longitude == 0 && p.Latitude == 0 {
				loc = pgtype.Point{
					Status: pgtype.Null,
				}
			} else {
				loc = pgtype.Point{
					P: pgtype.Vec2{
						X: p.Latitude,
						Y: p.Longitude,
					},
					Status: pgtype.Present,
				}
			}
			var combinedKey string
			if p.ProvinceState == "" {
				combinedKey = p.CountryRegion
			} else {
				combinedKey = fmt.Sprintf("%s, %s", p.ProvinceState, p.CountryRegion)
			}
			return []interface{}{p.Id, p.IsoCode, p.FipsId, p.Admin, p.ProvinceState, p.CountryRegion, loc, combinedKey, p.Population}, nil
		}))

	if err != nil {
		return err
	}

	return nil
}

type Number struct {
	current int
	change  int
}

func normalize(m map[string]string) (map[time.Time]Number, error) {
	layout := "1/2/06"
	normalized := make(map[time.Time]Number)
	for key, el := range m {
		t, err := time.Parse(layout, key)
		if err != nil {
			continue
		}
		i, err := strconv.Atoi(el)
		if err != nil {
			continue
		}
		normalized[t] = Number{current: i}
	}

	for key, el := range normalized {
		prev, found := normalized[key.AddDate(0, 0, -1)]
		if !found {
			prev = Number{}
		}
		normalized[key] = Number{
			current: el.current,
			change:  el.current - prev.current,
		}
	}

	return normalized, nil
}

func (ingestor *ingestorService) ingestCaseNumbers() error {
	confirmedGlobal, err := ingestor.gh.getNumbers(CONFIRMED, GLOBAL)
	if err != nil {
		return err
	}
	deathsGlobal, err := ingestor.gh.getNumbers(DEATHS, GLOBAL)
	if err != nil {
		return err
	}
	for i := range confirmedGlobal {
		if confirmedGlobal[i]["Country/Region"] != deathsGlobal[i]["Country/Region"] || confirmedGlobal[i]["Province/State"] != deathsGlobal[i]["Province/State"] {
			return fmt.Errorf("mismatch at index %d", i)
		}
	}

	//fmt.Println(confirmedGlobal[0])
	//fmt.Printf("%d, %d\n", len(confirmedGlobal), len(deathsGlobal))

	for i := range confirmedGlobal {
		var combinedKey string
		if ps := confirmedGlobal[i]["Province/State"]; ps != "" {
			combinedKey = fmt.Sprintf("%s, %s", ps, confirmedGlobal[i]["Country/Region"])
		} else {
			combinedKey = confirmedGlobal[i]["Country/Region"]
		}
		var id int
		err = ingestor.conn.QueryRow(context.Background(), "select id from places where combined_key = $1", combinedKey).Scan(&id)
		if err != nil {
			continue
		}
		fmt.Printf("%s, %d\n\n", combinedKey, id)
		_, err := normalize(confirmedGlobal[i])
		if err != nil {
			return err
		}
	//normalizedD, err := normalize(deathsGlobal[i])
	//if err != nil {
	//	return err
	//}

	}

	return nil
}
