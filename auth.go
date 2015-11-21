package metrik

//AuthRequest represents an authorization request. Credentials are passed through HTTP Basic Auth headers.
type AuthRequest struct {
	User     string
	Password string
	Metrics  []string
	Tags     []string
}

//AuthError represents an authentication error.
type AuthError struct {
	HTTPStatus int    //HTTP status to return (eg. 401)
	Message    string //Message that will be returned with the error
}

//AuthProvider represents an authentication provider. An authentication provider
//is a hook that authenticates incoming requests.
type AuthProvider interface {
	Authorize(*AuthRequest) (bool, *AuthError)
}

//type OpenAPI represents an API with no authorization or authentication (i.e. every request)
type openAPI struct{}

func (o openAPI) Authorize(a *AuthRequest) (bool, *AuthError) {
	return true, nil
}
