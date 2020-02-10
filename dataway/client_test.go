package dataway

import (
	"testing"
	"time"
)

func TestMarshalTags(t *testing.T) {

	m := map[string]string{
		"hostname": "MacBook-Pro",
		"ip":       "127.0.0.1",
	}

	t.Log("map size :", len(m))
	t.Log(marshalTags(m))

}

func TestMarshalFields(t *testing.T) {

	m := map[string]interface{}{
		"string": "hello,world",
		"int":    123456,
		"float":  22.33,
		"bytes":  []byte{65, 66, 67, 68},
		"struct": struct {
			n string
			i int
		}{
			n: "struct_n",
			i: 666666,
		},
		"escape": `  , , " " ' = = `,
	}

	t.Log("map size :", len(m))
	t.Log(marshalFields(m))

}

func TestDataway(t *testing.T) {

	c, err := New(&Option{
		DatawayHost:     "http://192.168.0.11:8888",
		AccessKey:       "ak123",
		SecretKey:       "sk123",
		X_TraceId:       "go_sdk_traceid",
		X_Datakit_UUID:  "go_sdk_uuid",
		X_Version:       "go_sdk_version",
		UserAgent:       "go_sdk_agent",
		NotGzipCompress: false,
	})
	if err != nil {
		panic(err)
	}

	var pts []*Point
	pts = append(pts, &Point{
		Name: "go_sdk_measurement",
		Tags: map[string]string{
			"hostname": "MacBook-Pro",
			"ip":       "127.0.0.1",
		},
		Fields: map[string]interface{}{
			"string": "hello,world",
			"int":    123456,
			"float":  22.33,
			"bytes":  []byte{41, 42, 43, 44},
			"struct": struct {
				n string
				i int
			}{
				n: "struct_n",
				i: 666666,
			},
			"escape": `  , , " " ' = = `,
		},
		Time: time.Now(),
	})

	pts = append(pts, &Point{
		Name: "go_sdk_measurement",
		Tags: map[string]string{
			"hostname": "MacBook-Pro",
			"ip":       "127.0.0.1",
		},
		Fields: map[string]interface{}{
			"string": "hello,world",
			"int":    123456,
			"float":  22.33,
			"bytes":  []byte{41, 42, 43, 44},
			"struct": struct {
				n string
				i int
			}{
				n: "struct_n",
				i: 666666,
			},
			"escape": `  , , " " ' = = `,
		},
		Time: time.Now(),
	})

	resp, err := c.Upload(&URLParam{}, pts, true)
	if err != nil {
		panic(err)
	}

	t.Log(resp.Status)
}
