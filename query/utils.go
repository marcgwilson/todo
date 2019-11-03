package query

import (
	"net/url"
	"strconv"
)

func CopyValues(q url.Values) url.Values {
	c := url.Values{}
	for k, v := range q {
		w := make([]string, len(v))
		copy(w, v)
		c[k] = w
	}
	return c
}

func GetQueryPage(q url.Values) int {
	if vals, ok := q["page"]; ok {
		if page, err := strconv.Atoi(vals[0]); err == nil {
			return page
		}
	}
	return 1
}
