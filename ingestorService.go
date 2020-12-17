package main

import (
	"context"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type ingestorService struct {
	conn *pgx.Conn
	gh   *githubFetcher
}

func (i *ingestorService) ingestPlaces() error {
	places, err := i.gh.getPlaces()
	if err != nil {
		return err
	}

	_, err = i.conn.CopyFrom(
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
			return []interface{}{p.Id, p.IsoCode, p.FipsId, p.Admin, p.ProvinceState, p.CountryRegion, loc, p.CombinedKey, p.Population}, nil
		}))

	if err != nil {
		return err
	}

	return nil
}
