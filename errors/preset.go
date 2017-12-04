package errors

// Types mapping to HTTP status 4xx and 5xx codes.
var (
	BadRequest                   = NewType(400) // RFC 7231, 6.5.1
	Unauthorized                 = NewType(401) // RFC 7235, 3.1
	PaymentRequired              = NewType(402) // RFC 7231, 6.5.2
	Forbidden                    = NewType(403) // RFC 7231, 6.5.3
	NotFound                     = NewType(404) // RFC 7231, 6.5.4
	MethodNotAllowed             = NewType(405) // RFC 7231, 6.5.5
	NotAcceptable                = NewType(406) // RFC 7231, 6.5.6
	ProxyAuthRequired            = NewType(407) // RFC 7235, 3.2
	RequestTimeout               = NewType(408) // RFC 7231, 6.5.7
	Conflict                     = NewType(409) // RFC 7231, 6.5.8
	Gone                         = NewType(410) // RFC 7231, 6.5.9
	LengthRequired               = NewType(411) // RFC 7231, 6.5.10
	PreconditionFailed           = NewType(412) // RFC 7232, 4.2
	RequestEntityTooLarge        = NewType(413) // RFC 7231, 6.5.11
	RequestURITooLong            = NewType(414) // RFC 7231, 6.5.12
	UnsupportedMediaType         = NewType(415) // RFC 7231, 6.5.13
	RequestedRangeNotSatisfiable = NewType(416) // RFC 7233, 4.4
	ExpectationFailed            = NewType(417) // RFC 7231, 6.5.14
	Teapot                       = NewType(418) // RFC 7168, 2.3.3
	UnprocessableEntity          = NewType(422) // RFC 4918, 11.2
	Locked                       = NewType(423) // RFC 4918, 11.3
	FailedDependency             = NewType(424) // RFC 4918, 11.4
	UpgradeRequired              = NewType(426) // RFC 7231, 6.5.15
	PreconditionRequired         = NewType(428) // RFC 6585, 3
	TooManyRequests              = NewType(429) // RFC 6585, 4
	RequestHeaderFieldsTooLarge  = NewType(431) // RFC 6585, 5
	UnavailableForLegalReasons   = NewType(451) // RFC 7725, 3

	InternalServerError           = NewType(500) // RFC 7231, 6.6.1
	NotImplemented                = NewType(501) // RFC 7231, 6.6.2
	BadGateway                    = NewType(502) // RFC 7231, 6.6.3
	ServiceUnavailable            = NewType(503) // RFC 7231, 6.6.4
	GatewayTimeout                = NewType(504) // RFC 7231, 6.6.5
	HTTPVersionNotSupported       = NewType(505) // RFC 7231, 6.6.6
	VariantAlsoNegotiates         = NewType(506) // RFC 2295, 8.1
	InsufficientStorage           = NewType(507) // RFC 4918, 11.5
	LoopDetected                  = NewType(508) // RFC 5842, 7.2
	NotExtended                   = NewType(510) // RFC 2774, 7
	NetworkAuthenticationRequired = NewType(511) // RFC 6585, 6
)
