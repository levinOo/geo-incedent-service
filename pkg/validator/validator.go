package validator

import (
	"errors"
	"fmt"

	"github.com/levinOo/geo-incedent-service/internal/entity"
)

// Вспомогательная функция для валидации GeoJsonPolygon
func ValidatePolygon(area entity.GeoJsonPolygon) error {
	if area.Type != "Polygon" {
		return errors.New("invalid GeoJsonPolygon type")
	}

	if len(area.Coordinates) == 0 {
		return errors.New("area must contain at least 1 coordinate")
	}

	exterior := area.Coordinates[0]
	if len(exterior) < 4 {
		return errors.New("area must contain at least 4 coordinates")
	}

	first := exterior[0]
	last := exterior[len(exterior)-1]
	if first[0] != last[0] || first[1] != last[1] {
		return fmt.Errorf("polygon must be closed: first (%v) != last (%v)", first, last)
	}

	for _, ring := range area.Coordinates {
		for _, coord := range ring {
			lat, lon := coord[1], coord[0]
			if lat < -90 || lat > 90 {
				return fmt.Errorf("invalid latitude: %f", lat)
			}
			if lon < -180 || lon > 180 {
				return fmt.Errorf("invalid longitude: %f", lon)
			}
		}
	}

	return nil
}

func ValidateLocation(location entity.UserLocation) error {
	if location.Lat < -90 || location.Lat > 90 {
		return fmt.Errorf("invalid latitude: %f", location.Lat)
	}
	if location.Lon < -180 || location.Lon > 180 {
		return fmt.Errorf("invalid longitude: %f", location.Lon)
	}
	return nil
}
