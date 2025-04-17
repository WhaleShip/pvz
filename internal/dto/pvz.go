package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/gen"
)

type ReceptionWithProducts struct {
	Reception Reception
	Products  []gen.Product
}

type PVZWithReceptions struct {
	Pvz        gen.PVZ
	Receptions []ReceptionWithProducts
}

type Reception struct {
	Id            *uuid.UUID          `json:"id"`
	PvzId         uuid.UUID           `json:"pvzId"`
	DateTime      time.Time           `json:"dateTime"`
	CloseDateTime *time.Time          `json:"closeDateTime,omitempty"`
	Status        gen.ReceptionStatus `json:"status"`
}
