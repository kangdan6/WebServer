package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	Req *http.Request

	//如果用户用这个，就绕开了RespStatusCode、RespData， 就获取不到，
	//那么部分middleware就无法运作
	Resp http.ResponseWriter

	//为了在middleware里读写用的
	//缓存响应部分，这部分在最后刷新
	RespStatusCode int
	RespData       []byte

	PathParams map[string]string
	//缓存住query数据，避免重复解析
	queryValues url.Values
	//命中的路由
	MatchedRoute string
}

func (c *Context) RespJsonOK(val any) error {
	return c.RespJson(http.StatusOK, val)
}

func (c *Context) RespJson(status int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	// _, err = c.Resp.Write(data)
	c.RespStatusCode = status
	c.RespData = data
	return err
}

// 解决大多数人的需求
func (c *Context) BindJson(val any) error {
	//这里不检测也没关系，decoder.Decode会检测
	if c.Req.Body == nil {
		return errors.New("web: body 为nil")
	}

	//bs, _ := io.ReadAll(c.Req.Body)
	//return json.Unmarshal(bs, val)

	decoder := json.NewDecoder(c.Req.Body)
	// UseNumber causes the Decoder to unmarshal a number into an interface{} as a
	// Number instead of as a float64.
	//decoder.UseNumber()
	//json中如果有未知字段就会报错
	//decoder.DisallowUnknownFields()
	return decoder.Decode(val)
}

// From是有缓存的，比如调用ParseForm，里面会判断if r.PostForm == nil， 所以不用担心重复调用重复解析的问题
func (c *Context) FormValue(key string) (string, error) {
	if err := c.Req.ParseForm(); err != nil {
		return "", err
	}
	return c.Req.FormValue(key), nil

}

// Query和表单比起来，没有缓存，存在重复解析的问题
func (c *Context) QueryValue(key string) (string, error) {
	//这种调用区别不出来是真的有值，值为空字符串，还是没有值
	//return c.Req.Form.Get(key), nil
	if c.queryValues == nil {
		c.queryValues = c.Req.URL.Query()
	}
	vals, ok := c.queryValues[key]
	if !ok {
		return "", errors.New("web: key 不存在")
	}
	return vals[0], nil

}

func (c *Context) PathValue(key string) (string, error) {
	val, ok := c.PathParams[key]
	if !ok {
		return "", errors.New("web: key 不存在")
	}
	return val, nil
}

func (c *Context) PathValue1(key string) StringValue {
	val, ok := c.PathParams[key]
	if !ok {
		return StringValue{
			err: errors.New("web: key 不存在"),
		}
	}
	return StringValue{
		val: val,
	}
}

type StringValue struct {
	val string
	err error
}

func (s StringValue) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseInt(s.val, 10, 64)

}

// 不能用泛型
// func (s StringValue) To[T any]() (T, error) {
//
// }
