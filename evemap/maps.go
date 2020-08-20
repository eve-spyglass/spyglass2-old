package evemap

import (
	"encoding/json"
	"github.com/Crypta-Eve/spyglass2/logger"
	"time"
)

type (

	NewEden struct {
		Maps map[string]MapFormat
	}

	MapFormat struct {
		Name     string        `json:"name"`
		SafeName string        `json:"dname"`
		Systems  map[int]Coord `json:"coords"`
		Gates    []Link        `links`
	}

	Coord struct {
		X string `json:"x"`
		Y string `json:"y"`
	}

	Link struct {
		Source      int `json:"src"`
		Destination int `json:"dest"`
	}
)

func CreateNewEden() (*NewEden, error) {
	start := time.Now()
	var maps map[string]MapFormat
	err := json.Unmarshal([]byte(mapdump), &maps)
	if err != nil{
		return nil, err
	}

	neweden := NewEden{Maps: maps}


	logger.Log.WithField("time", time.Since(start)).Info("new eden created")
	return &neweden, nil
}
