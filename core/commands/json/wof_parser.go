package json

import (
	"github.com/valyala/fastjson"
)

type State int

const (
	NoMatch             State = iota
	LatitudeState       State = iota
	LongitudeState      State = iota
	PreferredNamesState State = iota
	VariantNamesState   State = iota
	CountryCodeState    State = iota
	NA                  State = iota
	PossiblyLatLng      State = iota
)

var PossiblyLatLngBytes = []byte("eduti")
var LatBytes = []byte("al")
var LngBytes = []byte("nol")
var CountryBytes = []byte("yrtnuoc")
var PreferredBytes = []byte("derreferp_x_")
var VariantBytes = []byte("tnairav_x_")

func determineStateFSM(key []byte, skipLat bool, skipLong bool, skipCountry bool) (State, string) {
	length := len(key)
	state := NA
	scan := 0
	for i := length - 1; i >= 0; i-- {
		curChar := key[i]
		switch state {
		case LatitudeState:
			if scan > 6 {
				return LatitudeState, ""
			} else if curChar != LatBytes[scan-6] {
				return NoMatch, ""
			}
		case LongitudeState:
			if scan > 7 {
				return LongitudeState, ""
			} else if curChar != LngBytes[scan-6] {
				return NoMatch, ""
			}
		case PreferredNamesState:
			if scan > 11 {
				if curChar == ':' {
					return PreferredNamesState, string(key[length-scan : length-12])
				}
			} else if curChar != PreferredBytes[scan] {
				return NoMatch, ""
			}
		case VariantNamesState:
			if scan > 9 {
				if curChar == ':' {
					return VariantNamesState, string(key[length-scan : length-10])
				}
			} else if curChar != VariantBytes[scan] {
				return NoMatch, ""
			}
		case CountryCodeState:
			if scan > 6 {
				return CountryCodeState, ""
			} else if curChar != CountryBytes[scan] {
				return NoMatch, ""
			}
		case PossiblyLatLng:
			if scan > 4 {
				if curChar == 'g' {
					if skipLong {
						return NoMatch, ""
					}
					state = LongitudeState
				} else if curChar == 't' {
					if skipLat {
						return NoMatch, ""
					}
					state = LatitudeState
				} else {
					return NoMatch, ""
				}
			} else if curChar != PossiblyLatLngBytes[scan] {
				return NoMatch, ""
			}
		case NA:
			switch curChar {
			case 'd':
				state = PreferredNamesState
			case 't':
				state = VariantNamesState
			case 'e':
				state = PossiblyLatLng
			case 'y':
				if skipCountry {
					return NoMatch, ""
				}
				state = CountryCodeState
			default:
				return NoMatch, ""
			}
		}
		scan++
	}
	return NoMatch, ""
}

type jsonHolder struct {
	latitude       float64
	longitude      float64
	countryCode    string
	preferredNames map[string]string
	variantNames   map[string]string
	metadata       map[string]string
}

func (jc *JsonConverter) visitor(key []byte, v *fastjson.Value) {

	state, lang := determineStateFSM(key, jc.content.latitude != 0, jc.content.longitude != 0, jc.content.countryCode != "")
	switch state {
	case NoMatch:
		if v.Type() == fastjson.TypeString {
			strKey := string(key)
			jc.content.metadata[strKey] = string(v.GetStringBytes())
		}
	case LatitudeState:
		jc.content.latitude = v.GetFloat64()
	case LongitudeState:
		jc.content.longitude = v.GetFloat64()
	case PreferredNamesState:
		array := v.GetArray()
		if len(array) > 0 {
			jc.content.preferredNames[lang] = string(v.GetArray()[0].GetStringBytes())
		}
	case VariantNamesState:
		array := v.GetArray()
		if len(array) > 0 {
			jc.content.variantNames[lang] = string(v.GetArray()[0].GetStringBytes())
		}
	case CountryCodeState:
		jc.content.countryCode = string(v.GetStringBytes())
	}
}

type JsonConverter struct {
	content *jsonHolder
}

func ParseJson(value *fastjson.Value) map[string]interface{} {
	content := make(map[string]interface{})
	content["wof_id"] = value.GetInt64("id")
	props := value.GetObject("properties")
	req := jsonHolder{}
	req.preferredNames = make(map[string]string, props.Len())
	req.variantNames = make(map[string]string, props.Len())
	req.metadata = make(map[string]string, props.Len())

	get := props.Get("wof:hierarchy")
	if get != nil {
		hierarchy, err := get.Array()
		if err == nil && len(hierarchy) > 0 {
			content["continent_id"] = hierarchy[0].GetInt64("continent_id")
			content["country_id"] = hierarchy[0].GetInt64("country_id")
			content["locality_id"] = hierarchy[0].GetInt64("locality_id")
			content["macroregion_id"] = hierarchy[0].GetInt64("macroregion_id")
			content["region_id"] = hierarchy[0].GetInt64("region_id")
			content["county_id"] = hierarchy[0].GetInt64("county_id")
		}
	}
	jc := JsonConverter{
		&req,
	}
	props.Visit(jc.visitor)
	content["preferred_names"] = req.preferredNames
	content["variant_names"] = req.variantNames
	content["metadata"] = req.metadata
	content["country_code"] = req.countryCode
	if req.longitude != 0 {
		content["longitude"] = req.longitude
	}
	if req.latitude != 0 {
		content["latitude"] = req.latitude
	}
	return content
}
