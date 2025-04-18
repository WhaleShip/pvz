package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/gen/oapi"
)

type ReceptionWithProducts struct {
	Reception Reception
	Products  []oapi.Product
}

type PVZWithReceptions struct {
	Pvz        oapi.PVZ
	Receptions []ReceptionWithProducts
}

type Reception struct {
	Id            *uuid.UUID           `json:"id"`
	PvzId         uuid.UUID            `json:"pvzId"`
	DateTime      time.Time            `json:"dateTime"`
	CloseDateTime *time.Time           `json:"closeDateTime,omitempty"`
	Status        oapi.ReceptionStatus `json:"status"`
}
