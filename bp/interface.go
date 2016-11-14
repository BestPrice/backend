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
	Categories() ([]Category, error)
	Chainstores() ([]Chainstore, error)
	Stores(chainstore, district, region string) ([]Store, error)
	Products(category *ID, phrase string) ([]Product, error)
	Shop() (Shop, error)
}
