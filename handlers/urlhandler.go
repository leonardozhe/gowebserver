package handlers

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime/debug"
	"strings"

	"github.com/leonardozhe/gowebserver/models"
	"github.com/leonardozhe/gowebserver/utils"
	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

var _24_HOURS = 3600 * 24

// ServeURLHandler serve the url with choices
func ServeURLHandler(ctx *gin.Context) {
	request := ctx.Request
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered in f", r, string(debug.Stack()))
		}
	}()

	request.ParseForm()
	_ip := net.ParseIP(utils.GetGEOIP(request))
	ip := _ip.String()
	if models.InSimpCampIPBlackList(ip) {
		ctx.String(http.StatusNoContent, "")
		return
	}

	uid := ctx.Query("uid")
	cb := ctx.Query("cb")
	if uid == "" || cb == "" || strings.ContainsAny(cb, "#%{") {
		models.SaveSimpCampIP(ip)
		ctx.String(http.StatusNoContent, "")
		return
	}

	camp, ok := models.GetSimCamp(uid)
	if !ok {
		// @TBD
		// ipBlackListRecorder(ip)
		// ctx.String(http.StatusNoContent, "")
		ctx.String(http.StatusNotFound, "")
		return
	}

	ipInfos := utils.Get_CountryShort_UsageType(ip)
	if !models.IsWhiteIP(ip) &&
		(ipInfos.Usagetype == "DCH" || ipInfos.Country_short != camp.Geo) {
		models.SaveSimpCampIP(ip)
		ctx.String(http.StatusNoContent, "")
		return
	}

	if !handleCB(uid, cb) {
		// @TBD
		// why we have to flag this ip in blacklist?
		models.SaveSimpCampIP(ip)
		ctx.String(http.StatusNoContent, "")
		return
	}

	pid := ctx.Query("pid")
	newUrl := fmt.Sprintf("%s?pid=%s", camp.Url, pid)
	log.Info("newUrl, ", newUrl)

	doReverseProxy(ctx, newUrl)
	return
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func doReverseProxy(ctx *gin.Context, targetUrl string) {
	u, _ := url.Parse(targetUrl)

	//targetQuery := u.RawQuery
	director := func(req *http.Request) {
		headers := []string{
			"User-Agent",
			"Cookie",
			"Content-Type",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Referer",
		}
		for _, v := range headers {
			req.Header.Set(v, ctx.Request.Header.Get(v))
		}

		req.Host = u.Host

		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		req.URL.Path = u.Path
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(ctx.Writer, ctx.Request)
}

func getPage(originalReq *http.Request, url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	headers := []string{
		"User-Agent",
		"Cookie",
		"Content-Type",
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"Referer",
	}
	for _, v := range headers {
		req.Header.Set(v, originalReq.Header.Get(v))
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func handleCB(uid string, cb string) bool {
	conn := models.GetRedisConn()
	if conn == nil {
		log.Error("IsCBExist", models.ErrorRedisNotAvailable)
		return false
	}
	defer conn.Close()

	// cb shall not contains any of {'#', '%', '{'}
	key := uid + "##" + cb
	res, err := redis.Int(conn.Do("EXISTS", key))
	if err != nil {
		log.Error("IsCBExist", err)
		return false
	}

	if res == 0 {
		conn.Send("MULTI")
		conn.Send("SETEX", key, _24_HOURS, "1")

		if _, err := conn.Do("EXEC"); err != nil {
			log.Error("IsCBExist", err)
			return false
		}

		return true
	}

	return false
}
