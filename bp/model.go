package bp

type User struct {
}

type Category struct {
	ID            ID         `json:"id_category"`
	IDParent      ID         `json:"-"`
	Name          string     `json:"name"`
	Subcategories []Category `json:"subcategories,omitempty"`
}

type Chainstore struct {
	ID   ID     `json:"id_chain_store"`
	Name string `json:"name"`
}

type Store struct {
	IDChainstore ID     `json:"id_chain_store"`
	Name         string `json:"name"`
	City         string `json:"city"`
	Street       string `json:"street_and_nr"`
	District     string `json:"district"`
	Region       string `json:"region"`
	// Coordinates  string `json:coordinates`
}

type Product struct {
	ID                 ID             `json:"id_product"`
	Name               string         `json:"name"`
	Weight             JsonNullInt64  `json:"weigth"`
	Volume             JsonNullInt64  `json:"volume"`
	PriceDescription   JsonNullString `json:"price_description"`
	DecimalPossibility JsonNullBool   `json:"decimal_possibility"`
	Brand              Brand          `json:"brand"`

	Rank int `json:"-"`
}

type Brand struct {
	ID   ID     `json:"id_brand"`
	Name string `json:"name"`
	// IDChainstore ID     `json:"id_chain_store"`
}

type Shop struct{}
