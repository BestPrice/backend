package bp

// Client creates a connection to the services.
type Client interface {
	Service() Service
}

type Service interface {
	Categories() ([]Category, error)
	Chainstores() ([]Chainstore, error)
	Stores(chainstore, district, region string) ([]Store, error)
	Products(category *ID, phrase string) ([]Product, error)
	Shop() (Shop, error)
}
