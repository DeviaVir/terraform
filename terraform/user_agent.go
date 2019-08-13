package terraform

import (
	"github.com/DeviaVir/terraform/httpclient"
)

// Generate a UserAgent string
//
// Deprecated: Use httpclient.UserAgentString if you are setting your
// own User-Agent header.
func UserAgentString() string {
	return httpclient.UserAgentString()
}
