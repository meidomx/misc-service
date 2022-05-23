package dratelimiter

type RateTokenRequest struct {
	ResourceId string
	MaxCount   int32

	RequestId    string
	ClientId     string
	ReqTimestamp int64
}

type RateTokenResponse struct {
	AcquiredCount int32 // min(request.MaxCount, server available count, server configuration of max count of a single client)

	TtlInMillis int32 // server side ttl which tells the client to keep at most the certain amount time of the tokens

	ResourceId   string
	RequestId    string
	ClientId     string
	ReqTimestamp int64 // client side will check the availability of the token and discord invalid tokens if the time exceeded when client received the response
}
