package tileset

import (
	"encoding/json"
	"io"
	"os"
)

type LayerDataProperty struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type LayerData = struct {
	Data       []float64 `json:"data"`
	Name       string    `json:"name"`
	Visible    bool      `json:"visible"`
	Properties []LayerDataProperty
}

func LayerGetProp(ld LayerData, name string) int {
	for _, p := range ld.Properties {
		if p.Name == name {
			return p.Value
		}

	}
	return -1
}

type TileMapData = struct {
	TileWidth  int         `json:"tilewidth"`
	TileHeight int         `json:"tileheight"`
	Width      int         `json:"width"`
	Height     int         `json:"height"`
	Layers     []LayerData `json:"layers"`
}

func ParseMap(filepath string) (TileMapData, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return TileMapData{}, err
	}
	defer jsonFile.Close()

	var res TileMapData
	buffer, err := io.ReadAll(jsonFile)
	if err != nil {
		return TileMapData{}, err
	}
	err = json.Unmarshal(buffer, &res)
	if err != nil {
		return TileMapData{}, err
	}
	return res, nil
}
