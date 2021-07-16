package json

import (
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fastjson"
	"testing"
)

func TestInit(t *testing.T) {
	v, err := fastjson.Parse(TestData)
	assert.Nil(t, err)
	req := ParseJson(v)
	assert.Equal(t, int64(101750367), req["id"])
	assert.Equal(t, 51.500526, req["latitude"])
	assert.Equal(t, -0.109401, req["longitude"])
	assert.Equal(t, int64(102191581), req["continent_id"])
	preferredNames := req["preferred_names"].(map[string]string)
	assert.NotEmpty(t, preferredNames)
	variantNames := req["variant_names"].(map[string]string)
	assert.NotEmpty(t, req["variant_names"])
	assert.Equal(t, "London", preferredNames["eng"])
	assert.Equal(t, "Lodoni", preferredNames["fij"])
	assert.Equal(t, "London", preferredNames["eng_ca"])
	assert.Equal(t, "LON", variantNames["eng"])
	assert.Equal(t, "GB", req["country_code"])
	assert.NotEmpty(t, req["metadata"])
}
