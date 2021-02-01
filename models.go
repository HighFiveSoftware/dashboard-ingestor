package main

import "time"

type Case struct {
	CountryId int
	Confirmed int
	Deaths int
	Recovered int
	Date time.Time
}