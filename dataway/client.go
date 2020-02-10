// Package dataway implements functions for upload dataway service.

package dataway

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const _URL_PATH = "/v1/write/metrics?template=%s&token=%s&shortrp=%s"

type (
	Client struct {
		opt *Option
		url string
	}

	Option struct {
		DatawayHost string
		AccessKey   string
		SecretKey   string

		X_TraceId      string
		X_Datakit_UUID string
		X_Version      string
		UserAgent      string

		NotGzipCompress bool
	}

	Point struct {
		Name   string
		Tags   map[string]string
		Fields map[string]interface{}
		Time   time.Time
	}

	URLParam struct {
		Route   string
		Token   string
		Shortrp string
	}
)

func New(opt *Option) (*Client, error) {

	if opt.DatawayHost == "" {
		return nil, errors.New("invaild option, dataway host is empty")
	}

	return &Client{
		opt: opt,
		url: opt.DatawayHost + _URL_PATH,
	}, nil
}

func (c *Client) Upload(param *URLParam, pts []*Point, auth ...bool) (*http.Response, error) {
	var err error

	client := &http.Client{}

	local, err := time.LoadLocation("GMT")
	if err != nil {
		return nil, err
	}
	date := time.Now().In(local).Format(time.RFC1123)

	body := PointsToBytes(pts)
	// default use gzip compress
	if !c.opt.NotGzipCompress {
		body, err = gzipCompress(body)
		if err != nil {
			return nil, err
		}
	}

	requ, err := http.NewRequest("POST", fmt.Sprintf(c.url, param.Route, param.Token, param.Shortrp), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	requ.Header.Add("X-Trace-Id", c.opt.X_TraceId)
	requ.Header.Add("X-Datakit-UUID", c.opt.X_Datakit_UUID)
	requ.Header.Add("X-Version", c.opt.X_Version)
	requ.Header.Add("User-Agent", c.opt.UserAgent)
	requ.Header.Add("Content-Type", "text/plain")
	requ.Header.Add("Content-Length", strconv.Itoa(len(body)))
	requ.Header.Add("Date", date)

	if !c.opt.NotGzipCompress {
		requ.Header.Add("Content-Encoding", "gzip")
	}

	if len(auth) > 0 && auth[0] {
		requ.Header.Add("Authorization", c.signature(body, date))
	}

	return client.Do(requ)
}

func (c *Client) signature(body []byte, date string) string {
	bm := md5.Sum(body)

	var b bytes.Buffer
	b.WriteString("POST")
	b.WriteString("\n")
	b.WriteString(base64.StdEncoding.EncodeToString(bm[:]))
	b.WriteString("\n")
	b.WriteString("text/plain")
	b.WriteString("\n")
	b.WriteString(date)

	s := b.String()

	hm := hmac.New(func() hash.Hash { return sha1.New() }, []byte(c.opt.AccessKey))
	io.WriteString(hm, s)

	sg := base64.StdEncoding.EncodeToString(hm.Sum(nil))

	return "DWAY " + c.opt.SecretKey + ":" + sg
}

func PointsToBytes(pts []*Point) []byte {
	var b bytes.Buffer

	for _, p := range pts {
		if p == nil {
			continue
		}

		if len(p.Fields) == 0 {
			continue
		}

		b.WriteString(escape(p.Name))
		b.WriteString(",")
		b.WriteString(marshalTags(p.Tags))
		b.WriteString(" ")
		b.WriteString(marshalFields(p.Fields))
		b.WriteString(" ")
		b.WriteString(strconv.Itoa(int(p.Time.UnixNano())))
		b.WriteString("\n")
	}

	return b.Bytes()
}

var (
	_ESCAPE_CHAR = [...]string{",", " ", `"`, "="}

	_ESCAPE_REPLACE = [...]string{`\,`, `\ `, `\"`, `\=`}
)

func escape(s string) string {

	// FIXME: ...
	for k, v := range _ESCAPE_CHAR {
		s = strings.Replace(s, v, _ESCAPE_REPLACE[k], -1)
	}
	return s
}

func marshalTags(t map[string]string) string {
	var keys = make([]string, 0, len(t))
	var b bytes.Buffer

	for k, _ := range t {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for k, v := range keys {
		b.WriteString(escape(v))
		b.WriteString("=")
		b.WriteString(escape(t[v]))
		if k != len(keys)-1 {
			b.WriteString(",")
		}
	}

	return b.String()
}

func marshalFields(f map[string]interface{}) string {
	var keys = make([]string, 0, len(f))
	var b bytes.Buffer

	for k, _ := range f {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for k, v := range keys {
		b.WriteString(escape(v))
		b.WriteString("=")

		// fields value
		vv := f[v]
		switch vv.(type) {
		case uint:
			b.WriteString(strconv.Itoa(int(vv.(uint))))
			b.WriteString("i")
		case uint8:
			b.WriteString(strconv.Itoa(int(vv.(uint8))))
			b.WriteString("i")
		case uint16:
			b.WriteString(strconv.Itoa(int(vv.(uint16))))
			b.WriteString("i")
		case uint32:
			b.WriteString(strconv.Itoa(int(vv.(uint32))))
			b.WriteString("i")
		case uint64:
			// FIXME: maybe integer overflow
			b.WriteString(strconv.Itoa(int(vv.(uint64))))
			b.WriteString("i")
		case int:
			b.WriteString(strconv.Itoa(int(vv.(int))))
			b.WriteString("i")
		case int8:
			b.WriteString(strconv.Itoa(int(vv.(int8))))
			b.WriteString("i")
		case int16:
			b.WriteString(strconv.Itoa(int(vv.(int16))))
			b.WriteString("i")
		case int32:
			b.WriteString(strconv.Itoa(int(vv.(int32))))
			b.WriteString("i")
		case int64:
			b.WriteString(strconv.Itoa(int(vv.(int64))))
			b.WriteString("i")
		case float32:
			b.WriteString(fmt.Sprintf("%g", vv.(float32)))
		case float64:
			b.WriteString(fmt.Sprintf("%g", vv.(float64)))
		case string:
			b.WriteString(`"`)
			b.WriteString(escape(vv.(string)))
			b.WriteString(`"`)
		case []byte:
			b.WriteString(`"`)
			b.WriteString(escape(string(vv.([]byte))))
			b.WriteString(`"`)
		default:
			b.WriteString(`"`)
			b.WriteString(escape(fmt.Sprintf("%v", vv)))
			b.WriteString(`"`)
		}

		if k != len(keys)-1 {
			b.WriteString(",")
		}
	}

	return b.String()

}

func gzipCompress(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	_, err := zw.Write(b)
	if err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
