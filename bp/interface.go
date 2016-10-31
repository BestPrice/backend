package bp

// Client creates a connection to the services.
type Client interface {
	Connect() Session
}

// Session represents authenticable connection to the services.
type Session interface {
	SetAuthToken(token string)
	Service() Service
}

type Authenticator interface {
	Authenticate(token string) error
}

type Service interface {
	Chainstores() ([]Chainstore, error)
	Stores(chainstore, district, region string) ([]Store, error)
	Categories() ([]Category, error)
	Products(query string) ([]Product, error)
	Shop() (Shop, error)
}
