package security

import "database/sql"

type GeoTravelService struct {
	db *sql.DB
}

func NewGeoTravelService(db *sql.DB) *GeoTravelService {
	return &GeoTravelService{db: db}
}

func (g *GeoTravelService) CheckTravelEvent(phone, currentCity, deviceHash string, handwritingSimilarity float64) bool {
	var primaryDevice, lastCity string
	query := `SELECT COALESCE(primary_device_hash, ''), COALESCE(current_location_city, '') FROM users WHERE phone = $1`
	err := g.db.QueryRow(query, phone).Scan(&primaryDevice, &lastCity)
	if err != nil {
		return true
	}

	if primaryDevice == "" {
		g.db.Exec(`UPDATE users SET primary_device_hash = $1, current_location_city = $2 WHERE phone = $3`, deviceHash, currentCity, phone)
		return true
	}

	if primaryDevice == deviceHash && handwritingSimilarity >= 0.65 {
		g.db.Exec(`UPDATE users SET current_location_city = $1 WHERE phone = $2`, currentCity, phone)
		return true
	}

	return handwritingSimilarity >= 0.70
}
