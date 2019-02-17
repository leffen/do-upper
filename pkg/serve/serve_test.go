package serve

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingResponse(t *testing.T) {
	resp := pingResponse{}
	resp.pingNum = 1
	js, err := resp.toJSON()
	assert.Nil(t, err)
	assert.True(t, len(js) > 0, "Expects som json back")
	assert.Contains(t, js, "ping_num")

	di := map[string]interface{}{}
	err = json.Unmarshal([]byte(js), &di)
	assert.Nil(t, err)

	fmt.Printf("JSON :%#v\n", di)

	assert.True(t, di["ping_num"].(float64) == 1.0)
	fmt.Printf("JSON :%s\n", js)

}
