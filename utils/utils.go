package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
	// log "github.com/cihub/seelog"
	// jwt "github.com/dgrijalva/jwt-go"
	// "github.com/leonardozhe/gowebserver/configfile"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// GenRandomString gen the random string
func GenRandomString(salt string) string {
	now := time.Now().String()
	tM := fmt.Sprintf("%s%s", now, getLocalIP())
	key := []byte(salt)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(tM))
	return hex.EncodeToString(h.Sum(nil))
}

func RandStringBytesMaskImprSrc(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "123445"
	}
	for _, i := range ifaces {
		if i.Name != "lo" {
			return i.HardwareAddr.String()
		}
	}
	return ""
}

// GetGEOIP To void the proxy problem
// WARNING!!
// To make the x-forwarded-for correctly return the ip of the user after a
// opera mini or uc web, we need to set the nginx as:
//
//    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
//
// After setting  the above, the r.Header.Get("X-FORWARDED-FOR")
// will return mulitple address. According to RFC  http://tools.ietf.org/html/rfc7239#page-6
// the first ip will always be the real ip.
//
// If you have set the option as:
//
//    proxy_set_header X-Forwarded-For $http_x_forwarded_for;
//
// the nginx will only return the IP as the opera proxy IP. Nothing more
// will be returned.
//
// For more information, please read:
//    https://dev.opera.com/articles/opera-mini-request-headers/
//    http://wjw7702.blog.51cto.com/5210820/1150225
//
// !!!!!!!!!!!!!!!!!!!!!!
// If you have change any of the code in this function, please test the
// below scenarios
//    1. With nginx setting up,
//        a. use opera mini to test the url redirect
//        b. use the change
//        c. use the ucweb
//    2. With only go as frontend,
//        a. use opera mini to test the url redirect
//        b. use the change
//        c. use the ucweb
// All the case above should record the correct IP of the clicks.
func GetGEOIP(r *http.Request) string {

	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		ips := strings.Split(ipProxy, ", ")
		ipLen := len(ips)
		for i := 0; i < ipLen; i++ {
			if !IsLocalNetworkIP(ips[i]) {
				return ips[i]
			}
		}
		return ips[0]
	}

	if r.RemoteAddr == "" {
		return ""
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func SignToken(id string, user string) (string, error) {
	return "", nil
	//token := jwt.New(jwt.GetSigningMethod("RS256"))

	//token.Claims["id"] = id
	//token.Claims["user"] = user
	//token.Claims["exp"] = time.Now().Unix() + 3600

	//tokenString, err := token.SignedString([]byte(configfile.PrivateKey))
	//if err != nil {
	//	log.Error("Failed to sign a new token, due to ", err)
	//	return "", err
	//}

	//return tokenString, nil
}

func VerifyToken(r *http.Request) (bool, string) {
	return true, ""
	// //	if r.Method != "POST" && r.Method != "PUT" && r.Method != "DELETE" &&
	// //		r.Method != "PATCH" {
	// //		return true
	// //	}

	// token, err := jwt.ParseFromRequest(r,
	// 	func(token *jwt.Token) (interface{}, error) {

	// 		return []byte(configfile.PublicKey), nil
	// 	})

	// if err != nil {
	// 	log.Warn("Failed to verify the token due to ", err)

	// 	return false, ""
	// }

	// if token == nil || !token.Valid {
	// 	return false, ""
	// }

	// return true, token.Claims["id"].(string)
}

// IsLocalNetworkIP check whether the ip is out of the public ip range
func IsLocalNetworkIP(ip string) bool {
	if strings.Contains(ip, ":") {
		return false
	}
	raw := strings.Split(ip, ".")

	if len(raw) < 2 {
		return true
	}

	f, s := raw[0], raw[1]
	if f == "10" {
		return true
	} else if f == "192" && s == "168" {
		return true
	} else if f == "172" && s <= "32" && s >= "16" {
		return true
	}
	return false
}
