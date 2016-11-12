package sql

import (
	"strings"

	"github.com/BestPrice/backend/bp"
)

var _ bp.Service = &Service{}

type Service struct {
	session *Session
}

func (s Service) Categories() ([]bp.Category, error) {
	rows, err := s.session.db.Query("SELECT * FROM product where price_description=''")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vals := make([]bp.Category, 0, 32)
	for rows.Next() {
		var p bp.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.IDBrand, &p.Weight, &p.Volume, &p.IDParentProduct, &p.PriceDescription, &p.DecimalPossibility); err != nil {
			return nil, err
		}
		vals = append(vals, bp.Category{
			ID:   p.ID,
			Name: p.Name,
		})
	}

	return vals, nil
}

func (s Service) Chainstores() ([]bp.Chainstore, error) {
	rows, err := s.session.db.Query("SELECT * FROM chain_store")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vals := make([]bp.Chainstore, 0, 32)
	for rows.Next() {
		var c bp.Chainstore
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		vals = append(vals, c)
	}

	return vals, nil
}

func (s Service) Products(name string) ([]bp.Product, error) {
	if name == "" {
		name = ".*"
	} else {
		name = strings.ToLower(name)
	}

	rows, err := s.session.db.Query("SELECT * FROM product where lower(product_name) similar to '%" + name + "%' and not price_description=''")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vals := make([]bp.Product, 0, 32)
	for rows.Next() {
		var p bp.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.IDBrand, &p.Weight, &p.Volume, &p.IDParentProduct, &p.PriceDescription, &p.DecimalPossibility); err != nil {
			return nil, err
		}
		vals = append(vals, p)
	}

	return vals, nil
}

func (s Service) Stores(chainstore, district, region string) ([]bp.Store, error) {
	rows, err := s.session.db.Query("SELECT * FROM store")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vals := make([]bp.Store, 0, 32)
	for rows.Next() {
		var s bp.Store
		if err := rows.Scan(&s.IDChainstore); err != nil {
			return nil, err
		}
		vals = append(vals, s)
	}

	return vals, nil
}

func (s Service) Shop() (bp.Shop, error) {
	panic("TODO: implement Shop")
}
