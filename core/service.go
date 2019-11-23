package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	GET  = "GET"
	POST = "POST"
)

func WSHandler(c *gin.Context) {
	ws := NewWebsocket(c)
	ws.WriteSafe(genResponseData("connected", nil))
	ws.Sub(GlobalTopicID)
}

func HTTPPubHandler(c *gin.Context) {
	var data interface{}
	var err error
	body, _ := c.GetRawData()
	clientMsg, err := UnmarshalClientMessage(body, getHub(c))
	if err == nil {
		data, err = clientMsg.Process(nil)
	}
	JONSWithSmartCode(c, data, err)
}

func JONSWithSmartCode(c *gin.Context, data interface{}, err error) {
	code := 200
	if err != nil {
		code = 500
	}
	c.JSON(code, composeReponse(data, err))
}

func HTTPGetHandler(c *gin.Context) {
	var data interface{}
	var err error

	topic := c.Query("topic")
	amount := c.DefaultQuery("amount", "10")
	if topic == "" {
		c.JSON(400, composeReponse(data, errors.New("missing topic")))
		return
	}
	amountN, err := strconv.Atoi(amount)
	if err != nil {
		c.JSON(400, composeReponse(data, err))
		return
	}

	dataBytes := BufGetN(topic, amountN)
	_data := []string{}
	for _, x := range dataBytes {
		s := string(x)
		_data = append(_data, s)
	}

	data = map[string]interface{}{
		"data":  _data,
		"count": len(_data),
	}
	c.JSON(200, composeReponse(data, err))
}

func StatusHandler(c *gin.Context) {
	c.JSON(200, composeReponse(getHub(c), nil))
}

func getHub(c *gin.Context) *Hub {
	return c.Request.Context().Value("hub").(*Hub)
}

func withHub(hub *Hub, fn func(*gin.Context)) func(*gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		c.Request = c.Request.WithContext(context.WithValue(ctx, "hub", hub))
		fn(c)
	}
}

func dynamicHub(fn func(*gin.Context)) func(*gin.Context) {
	return func(c *gin.Context) {
		u, _ := c.Get(gin.AuthUserKey)
		user := u.(string)
		hub := HUB_MAP.GetHub(user)
		withHub(hub, fn)(c)
	}
}

func ServeHub(listen string) {
	users := MustMapStr(LoadJSON("users.json").MustMap())
	fmt.Println("auth", ToJSONStr(users))

	log.Printf("serve http on %s", listen)
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "service is healthy",
			"source_code": "https://github.com/weaming/hub",
		})
	})

	index := r.Group("/")
	index.GET("/http", withHub(HUBPublic, HTTPGetHandler))
	index.POST("/http", withHub(HUBPublic, HTTPPubHandler))
	index.GET("/ws", withHub(HUBPublic, WSHandler))
	index.GET("/status", withHub(HUBPublic, StatusHandler))

	public := r.Group("/api/public")
	public.GET("/http", withHub(HUBPublic, HTTPGetHandler))
	public.POST("/http", withHub(HUBPublic, HTTPPubHandler))
	public.GET("/ws", withHub(HUBPublic, WSHandler))
	public.GET("/status", withHub(HUBPublic, StatusHandler))

	authShare := r.Group("/api/share", gin.BasicAuth(gin.Accounts(users)))
	authShare.GET("/http", withHub(HUBShare, HTTPGetHandler))
	authShare.POST("/http", withHub(HUBShare, HTTPPubHandler))
	authShare.GET("/ws", withHub(HUBShare, WSHandler))
	authShare.GET("/status", withHub(HUBShare, StatusHandler))

	authPrivate := r.Group("/api/private", gin.BasicAuth(gin.Accounts(users)))
	authPrivate.GET("/http", dynamicHub(HTTPGetHandler))
	authPrivate.POST("/http", dynamicHub(HTTPPubHandler))
	authPrivate.GET("/ws", dynamicHub(WSHandler))
	authPrivate.GET("/status", dynamicHub(StatusHandler))

	r.Run(listen)
}
