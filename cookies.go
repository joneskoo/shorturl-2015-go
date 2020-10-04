package shorturl

import "net/http"

// Cookie names used by shorturl service.
const (
	// alwaysPreviewCookie is the name of cookie for preference to
	// always show a preview page instead of immediately redirecting to
	// target. The value is "true" when enabled, otherwise ignored.
	alwaysPreviewCookie = "preview"
)

func alwaysPreviewPref(req *http.Request) bool {
	cookies := req.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == alwaysPreviewCookie && cookie.Value == "true" {
			return true
		}
	}
	return false
}
