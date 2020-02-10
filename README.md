# dataway-go-sdk

## 简介

dataway-go-sdk 是用 Go 开发的包，用于数据进行适配以行协议格式写入 DataWay 中，方便用户专注于自己业务逻辑开发，不需要数据有效性检查以及行协议格式的转换等。

## 安装

使用`go get`安装：
```
go get github.com/CloudCare/dataway-go-sdk
```

**需要 golang 1.13.5 及以上版本。**

## 使用方式

完整示例：
``` golang
package main

import "time"
import "github.com/CloudCare/dataway-go-sdk/dataway"

func main() {

	// 创建 dataway client
	c, err := dataway.New(&dataway.Option{
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

	// 组装要发送的数据
	var pts []*dataway.Point
	pts = append(pts, &dataway.Point{
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
		},
		Time: time.Now(),
	})

	// 上传，URLParam 为空时 Route 为'default'
	// Upload 第三个参数为可变参，为true时会添加'Authorization'验证，默认为false
	
	resp, err := c.Upload(&dataway.URLParam{}, pts, true)
	if err != nil {
		panic(err)
	}

	print(resp.Status)

}
```

## 参数解释

#### client 创建参数 Option

```
- DatawayHost		dataway 地址，`http://ip:port`
- AccessKey		验证公钥
- SecretKey		验证秘钥
- X_TraceId		trace id
- X_Datakit_UUID	当前客户端uuid，可直接生成uuid使用
- X_Version		版本
- UserAgent		
- NotGzipCompress	是否发送时进行gzip压缩，默认不压缩
```
#### 数据 point 格式

```
- Name		measurement名字，string
- Tags		tags组，map[string]string
- Fiedls	Fiedls组，map[string]interface{}
- Time		标准库time类型，当前 point 的时间
```

#### 上传参数 URLParam

```
- Route		上传到该路由，默认为`default`
- Token
- Shortrp
```
