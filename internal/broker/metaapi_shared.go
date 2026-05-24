package broker

import "strings"

// SharedMetaAPIToken is set from METAAPI_SHARED_TOKEN — clients only enter MT login/password/server.
var SharedMetaAPIToken string

// SetSharedMetaAPIToken configures operator-provided MetaAPI token for all MT connections.
func SetSharedMetaAPIToken(token string) {
	SharedMetaAPIToken = strings.TrimSpace(token)
}

// UsesSharedMetaAPIToken reports whether clients can skip entering their own MetaAPI token.
func UsesSharedMetaAPIToken() bool {
	return SharedMetaAPIToken != ""
}

// ApplySharedMetaAPIToken injects the platform token when the user did not supply one.
func ApplySharedMetaAPIToken(creds Credentials) Credentials {
	if creds == nil {
		creds = Credentials{}
	}
	if SharedMetaAPIToken == "" {
		return creds
	}
	if strings.TrimSpace(creds["metaapi_token"]) == "" &&
		strings.TrimSpace(creds["token"]) == "" &&
		strings.TrimSpace(creds["api_token"]) == "" {
		out := Credentials{}
		for k, v := range creds {
			out[k] = v
		}
		out["metaapi_token"] = SharedMetaAPIToken
		return out
	}
	return creds
}
