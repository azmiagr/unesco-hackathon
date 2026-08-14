package model

import "github.com/google/uuid"

const CityStatisticsDefaultKey = "default"

type GetCityStatisticsParam struct {
	CityStatisticsID uuid.UUID
	StatKey          string
}
