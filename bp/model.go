package bp

type User struct{}

type Chainstore struct {
	Id   int
	Name string
}

type Store struct {
	Id           int
	IdChainstore int
	Name         string
	City         string
	Street       string
	District     string
	Region       string
	Coordinates  string
}

type Category struct {
	Id       int    `json:"id"`
	IdParent int    `json:"-"`
	Name     string `json:"name"`

	Subcategories []Category `json:"subcategories,omitempty"`
}

type Product struct {
	Id                 int
	Name               string
	IdBrand            int
	IdCategory         int
	Weigth             int
	Volume             int
	IdParent           int
	PriceDescription   string
	DecimalPossibility bool
}

type ProductPrices struct {
	Id           int
	IdProduct    int
	IdChainstore int
	UnitPrice    string // Decimal(8,2)
}

type Brand struct {
	Id           int
	IdChainstore int
	Name         string
}

type Shop struct{}
