package mqtt_parsing

import (
	"log"
	"strings"
)

type TopicPath struct {
	Wildcard bool
	Split    []string
	Whole    string
}

func donttakelog() { log.Println("wut") }

//NewTopicPath creates a TopicPath object from the
//this function could be more efficient
func NewTopicPath(path string) (TopicPath, bool) {
	ff := func(r rune) bool {
		return r == '/'
	}
	path = strings.Trim(path, "/")
	split := strings.FieldsFunc(path, ff)
	//reorder or copy?
	spl := make([]string, 0, len(split))
	var tp TopicPath
	for i := 0; i < len(split); i++ {
		v := split[i]
		if len(v) == 0 || v == "" {
			continue
		}
		if (strings.Contains(v, "+") && 1 != len(v)) || (strings.Contains(v, "#") && 1 != len(v)) {
			return tp, false
		}

		if v == "#" {
			if i < len(split)-1 {
				return tp, false
			} else {
				tp.Wildcard = true
			}
		}

		if v == "+" {

			tp.Wildcard = true
		}

		spl = append(spl, v)
	}
	tp.Split = spl
	tp.Whole = path
	return tp, true
}

//so much garbage
func TrimPath(pth TopicPath, idx int, wildcard bool) TopicPath {
	t := pth.Split[:idx]
	s := strings.Join(t, "/")
	return TopicPath{
		Split:    t,
		Whole:    s,
		Wildcard: wildcard,
	}
}

func CutOffSystemKey(pth TopicPath, idx int, wildcard bool) (string, TopicPath) {
	appkey := pth.Split[0]
	return appkey, TopicPath{
		Split:    pth.Split[idx:],
		Whole:    strings.Join(pth.Split[idx:], "/"),
		Wildcard: wildcard,
	}
}
