package dto

import "github.com/whaleship/pvz/internal/gen"

type ReceptionWithProducts struct {
	Reception gen.Reception
	Products  []gen.Product
}

type PVZWithReceptions struct {
	Pvz        gen.PVZ
	Receptions []ReceptionWithProducts
}

type PVZResponse struct {
	PvzWithReceptions []PVZWithReceptions
}
