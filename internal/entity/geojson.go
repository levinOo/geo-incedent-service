package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type GeoJsonPolygon struct {
	Type        string        `json:"type" example:"Polygon"`
	Coordinates [][][]float64 `json:"coordinates"`
}

func (g GeoJsonPolygon) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GeoJsonPolygon) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, g)
	case string:
		return json.Unmarshal([]byte(v), g)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

func (g *GeoJsonPolygon) Contains(lat, lon float64) bool {
	if len(g.Coordinates) == 0 {
		return false
	}

	polygon := g.Coordinates[0]
	inside := false

	for i, j := 0, len(polygon)-1; i < len(polygon); j, i = i, i+1 {
		xi, yi := polygon[i][0], polygon[i][1]
		xj, yj := polygon[j][0], polygon[j][1]

		intersect := ((yi > lat) != (yj > lat)) &&
			(lon < (xj-xi)*(lat-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
	}

	return inside
}
