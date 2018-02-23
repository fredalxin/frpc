package selector

import (
	"time"
	"math"
	"net/url"
	"strconv"
	"math/rand"
	"context"
)

type geoSelector struct {
	servers   []*geoServer
	Latitude  float64
	Longitude float64
	r         *rand.Rand
}

type geoServer struct {
	Server    string
	Latitude  float64
	Longitude float64
}

func newGeoSelector(servers map[string]string, latitude, longitude float64) Selector {
	ss := createGeoServer(servers)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &geoSelector{servers: ss, Latitude: latitude, Longitude: longitude, r: r}
}

func (s geoSelector) Select(ctx context.Context, servicePath, serviceMethod string, args interface{}) string {
	if len(s.servers) == 0 {
		return ""
	}

	var server []string
	min := math.MaxFloat64
	for _, gs := range s.servers {
		d := getDistanceFrom(s.Latitude, s.Longitude, gs.Latitude, gs.Longitude)
		if d < min {
			server = []string{gs.Server}
			min = d
		} else if d == min {
			server = append(server, gs.Server)
		}
	}

	if len(server) == 1 {
		return server[0]
	}

	return server[s.r.Intn(len(server))]
}

func (s *geoSelector) UpdateServer(servers map[string]string) {
	ss := createGeoServer(servers)
	s.servers = ss
}

func createGeoServer(servers map[string]string) []*geoServer {
	var geoServers = make([]*geoServer, len(servers))

	for s, metadata := range servers {
		if v, err := url.ParseQuery(metadata); err == nil {
			latStr := v.Get("latitude")
			lonStr := v.Get("longitude")

			if latStr == "" || lonStr == "" {
				continue
			}

			lat, err := strconv.ParseFloat(latStr, 64)
			if err != nil {
				continue
			}
			lon, err := strconv.ParseFloat(lonStr, 64)
			if err != nil {
				continue
			}

			geoServers = append(geoServers, &geoServer{Server: s, Latitude: lat, Longitude: lon})

		}
	}

	return geoServers
}

//https://gist.github.com/cdipaolo/d3f8db3848278b49db68
func getDistanceFrom(lat1, lon1, lat2, lon2 float64) float64 {
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}
