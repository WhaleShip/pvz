package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/gen/oapi"
)

type ReceptionWithProducts struct {
	Reception Reception      `json:"receptions"`
	Products  []oapi.Product `json:"products"`
}

type PVZWithReceptions struct {
	Pvz        oapi.PVZ                `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

type Reception struct {
	Id            *uuid.UUID           `json:"id"`
	PvzId         uuid.UUID            `json:"pvzId"`
	DateTime      time.Time            `json:"dateTime"`
	CloseDateTime *time.Time           `json:"closeDateTime,omitempty"`
	Status        oapi.ReceptionStatus `json:"status"`
}
