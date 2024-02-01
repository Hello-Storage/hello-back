package constant

import (
	"fmt"
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
	APIKeyHeaderKey         = "api-key"
)

const LoginMessage = "Greetings from hello\nSign this message to log into hello\nnonce: "

func BuildLoginMessage(nonce string) []byte {
	return []byte(fmt.Sprintf("%s%s", LoginMessage, nonce))
}
