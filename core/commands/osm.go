package commands

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/meekyphotos/experive-cli/core/utils"
	"github.com/urfave/cli/v2"
	"strings"
)

type Feature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type OsmDownloader struct {
	regions []Region
}

type Region struct {
	Id           string
	Parent       string
	CountryCodes []string
	Name         string
	PbfLink      string
	Bz2Link      string
}

//go:embed index-v1-nogeom.json
var geofabrikJson []byte

func (o *OsmDownloader) Init() error {
	var data FeatureCollection
	err := json.Unmarshal(geofabrikJson, &data)
	if err != nil {
		return err
	}
	o.regions = make([]Region, len(data.Features))
	for _, f := range data.Features {
		id := f.Properties["id"].(string)
		var parent string
		if f.Properties["parent"] != nil {
			parent = f.Properties["parent"].(string)
		}
		//countryCodes := f.Properties["iso3166-1:alpha2"].([]interface{})
		name := f.Properties["name"].(string)
		urls := f.Properties["urls"].(map[string]interface{})
		o.regions = append(o.regions, Region{
			Id:      id,
			Parent:  parent,
			Name:    name,
			PbfLink: urls["pbf"].(string),
			Bz2Link: urls["bz2"].(string),
		})
	}
	return nil
}

func (o *OsmDownloader) createIndex(content string) error {
	var data FeatureCollection
	err := json.Unmarshal([]byte(content), &data)
	if err != nil {
		return err
	}

	return nil
}
func matches(region *Region, specification *Region) bool {
	if specification.Id == "" || strings.EqualFold(region.Id, specification.Id) {
		return true
	}
	if specification.Parent == "" || strings.EqualFold(region.Parent, specification.Parent) {
		return true
	}
	if specification.Name == "" || strings.EqualFold(region.Name, specification.Name) {
		return true
	}

	return false
}
func (o *OsmDownloader) List(context *cli.Context) error {
	region := context.Args().First()
	if region == "" {
		return errors.New("zone argument missing")
	}
	specification := Region{
		Id:     region,
		Parent: region,
		Name:   region,
	}
	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "Parent", "Country Code", "Name"})
	for _, r := range o.regions {

		if matches(&r, &specification) {
			t.AppendRow(table.Row{
				r.Id, r.Parent, r.CountryCodes, r.Name,
			})
		}
	}

	fmt.Println(t.Render())
	return nil
}

func (o *OsmDownloader) ById(region string) *Region {
	for _, r := range o.regions {
		if r.Id == region {
			return &r
		}
	}
	return nil
}

func (o *OsmDownloader) OsmDownload(context *cli.Context) error {
	format := context.String("format")
	zone := context.Args().First()
	if zone == "" {
		return errors.New("zone argument missing")
	}
	region := o.ById(zone)
	if region == nil {
		return errors.New("zone " + zone + " has not been found")
	}

	switch format {
	case "pbf":
		return utils.DownloadFile(region.Id+".osm.pbf", region.PbfLink)
	case "bz2":
		return utils.DownloadFile(region.Id+".osm.bz2", region.PbfLink)
	}
	return nil
}
