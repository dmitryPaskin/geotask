package geo

import (
	geo "github.com/kellydunn/golang-geo"
	"math/rand"
)

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type PolygonChecker interface {
	Contains(point Point) bool // проверить, находится ли точка внутри полигона
	Allowed() bool             // разрешено ли входить в полигон
	RandomPoint() Point        // сгенерировать случайную точку внутри полигона
}

type Polygon struct {
	polygon *geo.Polygon
	allowed bool
}

func NewPolygon(points []Point, allowed bool) *Polygon {
	// используем библиотеку golang-geo для создания полигона
	var geoPoints []*geo.Point
	for _, point := range points {
		geoPoints = append(geoPoints, geo.NewPoint(point.Lat, point.Lng))
	}

	return &Polygon{
		polygon: geo.NewPolygon(geoPoints),
		allowed: allowed,
	}
}

func (p *Polygon) Contains(point Point) bool {
	return p.polygon.Contains(geo.NewPoint(point.Lat, point.Lng))
}

func (p *Polygon) Allowed() bool {
	return p.allowed
}

func (p *Polygon) RandomPoint() Point {
	minLat := p.polygon.Points()[0].Lat()
	maxLat := p.polygon.Points()[0].Lat()
	minLng := p.polygon.Points()[0].Lng()
	maxLng := p.polygon.Points()[0].Lng()

	for _, point := range p.polygon.Points()[1:] {
		lat := point.Lat()
		lng := point.Lng()

		if lat < minLat {
			minLat = lat
		}
		if lat > maxLat {
			maxLat = lat
		}
		if lng < minLng {
			minLng = lng
		}
		if lng > maxLng {
			maxLng = lng
		}
	}

	randlat := rand.Float64()*(maxLat-minLat) + minLat
	randlng := rand.Float64()*(maxLng-minLng) + minLng

	return Point{Lat: randlat, Lng: randlng}
}

func CheckPointIsAllowed(point Point, allowedZone PolygonChecker, disabledZones []PolygonChecker) bool {
	// проверить, находится ли точка в разрешенной зоне
	if !allowedZone.Allowed() {
		return false
	}

	if !allowedZone.Contains(point) {
		return false
	}

	for _, disabledZone := range disabledZones {
		if disabledZone.Contains(point) || disabledZone.Allowed() {
			return false
		}
	}
	return true
}

func GetRandomAllowedLocation(allowedZone PolygonChecker, disabledZones []PolygonChecker) Point {
	var point Point
	for {
		point = allowedZone.RandomPoint()
		if CheckPointIsAllowed(point, allowedZone, disabledZones) {
			return point
		}
	}
}

func NewDisAllowedZone1() *Polygon {
	// добавить полигон с разрешенной зоной
	// полигоны лежат в /public/js/polygons.js
	points := []Point{
		{Lat: 60.9027, Lng: 31.3575},
		{Lat: 58.9001, Lng: 32.4158},
		{Lat: 59.8424, Lng: 30.4953},
		{Lat: 59.8896, Lng: 30.3736},
	}
	return NewPolygon(points, false)
}

func NewDisAllowedZone2() *Polygon {
	// добавить полигон с разрешенной зоной
	// полигоны лежат в /public/js/polygons.js
	points := []Point{
		{Lat: 61.2714, Lng: 31.2874},
		{Lat: 59.9604, Lng: 30.3413},
		{Lat: 61.0206, Lng: 31.3613},
		{Lat: 60.0151, Lng: 30.3752},
	}
	return NewPolygon(points, false)
}

func NewAllowedZone() *Polygon {
	// добавить полигон с разрешенной зоной
	// полигоны лежат в /public/js/polygons.js
	points := []Point{
		{Lat: 59.8337, Lng: 30.2997},
		{Lat: 59.2356, Lng: 30.2681},
		{Lat: 59.8868, Lng: 30.8263},
		{Lat: 59.8943, Lng: 30.7933},
	}
	return NewPolygon(points, true)
}
