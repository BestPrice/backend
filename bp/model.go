package bp

import (
	"errors"

	"github.com/shopspring/decimal"
)

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
	ID       ID              `json:"id_store"`
	CSName   JsonNullString  `json:"chain_store_name"`
	Name     JsonNullString  `json:"store_name"`
	City     JsonNullString  `json:"city"`
	Street   JsonNullString  `json:"street_and_nr"`
	District JsonNullString  `json:"district"`
	Region   JsonNullString  `json:"region"`
	Lat      decimal.Decimal `json:"latitude"`
	Lng      decimal.Decimal `json:"longitude"`
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

type ShopRequestProduct struct {
	ID    ID  `json:"id_product"`
	Count int `json:"count"`
}

type UserPreference struct {
	IDs       []ID `json:"id_chain_stores"`
	MaxStores int  `json:"max_stores"`
}

func (u *UserPreference) Contains(id ID) bool {
	if len(u.IDs) == 0 {
		return true
	}
	for _, cs := range u.IDs {
		if *cs.UUID == *id.UUID {
			return true
		}
	}
	return false
}

func (u *UserPreference) SetPrefered(IDs []ID) {
	u.IDs = IDs
}

type ShopRequest struct {
	Products       []ShopRequestProduct `json:"products"`
	UserPreference UserPreference       `json:"user_preference"`
}

func (s *ShopRequest) ProductCount(id ID) int {
	for _, p := range s.Products {
		if p.ID.String() == id.String() {
			return p.Count
		}
	}
	return 0
}

func (s *ShopRequest) Valid() error {
	if len(s.Products) == 0 {
		return errors.New("at least one product must be added")
	}
	if len(s.UserPreference.IDs) == 0 {
		return errors.New("at least one Chain Store must be set")
	}
	if s.UserPreference.MaxStores <= 0 {
		return errors.New("minimum one Chain Store must be set")
	}
	if s.UserPreference.MaxStores > len(s.UserPreference.IDs) {
		s.UserPreference.MaxStores = len(s.UserPreference.IDs)
	}
	return nil
}

type ShopProduct struct {
	ID           ID              `json:"id_product"`
	IDChainStore ID              `json:"-"`
	ChainStore   string          `json:"-"`
	Product      string          `json:"product_name"`
	Brand        string          `json:"brand_name"`
	Count        int             `json:"count"`
	PriceDesc    string          `json:"-"`
	Price        decimal.Decimal `json:"price"`
}

type ShopStore struct {
	ID             ID            `json:"-"`
	ChainStoreName string        `json:"chain_store_name"`
	Products       []ShopProduct `json:"products"`
	// PriceTotal     decimal.Decimal `json:"store_price_total"`
}

type Shop struct {
	Error string `json:"error,omitempty"`

	Stores     []ShopStore     `json:"stores,omitempty"`
	PriceTotal decimal.Decimal `json:"shop_price_total,omitempty"`
}
