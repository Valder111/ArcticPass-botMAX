package api

import "time"

type ErrorResponse struct {
	Error string `json:"error"`
}

type ObjectResponse struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	District    string    `json:"district"`
	Street      string    `json:"street"`
	HouseNumber string    `json:"house_number"`
	PostalCode  string    `json:"postal_code"`
	Levels      int       `json:"levels"`
	Material    string    `json:"material"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	OsmID       int64     `json:"osm_id"`
	OsmType     string    `json:"osm_type"`
	OsmBuilding string    `json:"osm_building"`
	Source      string    `json:"source"`
	QRCode      string    `json:"qr_code"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ApartmentResponse struct {
	YearOfEntry        int     `json:"year_of_entry"`
	Apartments         int     `json:"apartments"`
	TypeOfConstruction string  `json:"type_of_construction"`
	TypeOfWalls        string  `json:"type_of_walls"`
	SquareTotal        float64 `json:"square_total"`
	CadastralNumber    string  `json:"cadastral_number"`
	CompanyName        string  `json:"company_name"`
}
