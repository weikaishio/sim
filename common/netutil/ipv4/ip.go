package ipv4

import (
	"errors"
	"net"
	"strings"
)

var (
	ErrNoMatchedIP = errors.New("no matched IP")
	ErrAmbiguousIP = errors.New("ambiguous IP")
)

type patternT []string

func parsePattern(s string) patternT {
	return patternT(strings.Split(s, "."))
}

func (pat patternT) match(ip string) bool {
	vals := strings.Split(ip, ".")
	l := len(pat)
	if len(vals) < l {
		return false
	}
	for i := 0; i < l; i++ {
		if pat[i] != "" && pat[i] != "*" && pat[i] != vals[i] {
			return false
		}
	}
	return true
}

// (DOC): IPv4 匹配查找
//
// pattern 将被 "." 分割成数组 pats,
// 每个 IP 也被 "." 分割成数组 vals,
// 然后按如下规则进行匹配比较:
//
//	1. 如果 len(vals) < len(pats), 匹配失败
//	2. 如果 pats[i] 不等于 vals[i] 并且不为空也不为 "*",匹配失败
//	3. 其他情况为匹配成功
//
// 比如:
//
// `127.*` 匹配 IP(127.0.0.1) 成功
// `127.0.0.1.*` 匹配 IP(127.0.0.1) 失败
// `127.1.*` 匹配 IP(127.0.0.1) 失败

func FindOne(pattern string) (string, error) {
	if pattern == "0.0.0.0" {
		return "0.0.0.0", nil
	}
	foundIPs, err := FindAll(pattern)
	if err != nil {
		return "", err
	}
	if len(foundIPs) == 0 {
		return "", ErrNoMatchedIP
	} else if len(foundIPs) > 1 {
		return "", ErrAmbiguousIP
	}
	return foundIPs[0], nil
}

func FindAll(pattern string) ([]string, error) {
	pat := parsePattern(pattern)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var foundIPs []string
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue
		}
		if pat.match(ip.String()) {
			foundIPs = append(foundIPs, ip.String())
		}
	}
	return foundIPs, nil
}

//type IPGeoInfo struct {
//	Country string // 国家
//	Region  string // 省
//	City    string // 城市
//}
//
//var geoServiceAddr string
//
//func InitGeoService(serviceAddr string) {
//	geoServiceAddr = serviceAddr
//}
//
//// 获取IP地址信息
//// @param serviceAddr : IP 地理信息服务地址
//// @param ip          : 需要查询的IP
//func Geo(ip string) (info *IPGeoInfo, err error) {
//	resp, err := http.Get(geoServiceAddr + "?ip=" + ip)
//	if err != nil {
//		return
//	}
//	defer resp.Body.Close()
//	info = &IPGeoInfo{}
//	ret := &httputil.HTTPResponse{
//		Data: info,
//	}
//	err = json.NewDecoder(resp.Body).Decode(ret)
//	if err == nil && ret.Error != "" {
//		err = errors.New(ret.Error)
//	}
//	return
//}
