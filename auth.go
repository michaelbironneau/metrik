package metrik

//type AuthRequest represents an authorization request. Credentials are passed through HTTP Basic Auth headers.
type AuthRequest struct {
	User     string
	Password string
	Metrics  []string
	Tags     []string
}

type AuthError struct {
	HTTPStatus int
	Message    string
}

type AuthProvider interface {
	Authorize(*AuthRequest) (bool, *AuthError)
}

//type OpenAPI represents an API with no authorization or authentication (i.e. every request)
type openAPI struct{}

func (o openAPI) Authorize(a *AuthRequest) (bool, *AuthError) {
	return true, nil
}
