package query

import (
	"net/url"
	"strconv"
)

func NextPage(r *QueryParams, u *url.URL, count int64) string {
	if r.Offset().Offset()+r.Limit().Count() < count {
		u, _ := url.Parse(u.String())
		values := u.Query()
		values.Set("page", strconv.FormatInt(r.Offset().Page()+1, 10))
		queryString, _ := url.QueryUnescape(values.Encode())
		u.RawQuery = queryString
		return u.String()
	}
	return ""
}

func PrevPage(r *QueryParams, u *url.URL) string {
	if r.Offset().Page() > 1 {
		u, _ := url.Parse(u.String())
		values := u.Query()
		values.Set("page", strconv.FormatInt(r.Offset().Page()-1, 10))
		queryString, _ := url.QueryUnescape(values.Encode())
		u.RawQuery = queryString
		return u.String()
	}
	return ""
}
