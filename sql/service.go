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

	query := `
	WITH RECURSIVE nodes (id_product, product_name, id_parent_product)
	AS (
		SELECT p.id_product, p.product_name, p.id_parent_product
		FROM product p
		WHERE p.id_parent_product is NULL
		UNION ALL
		SELECT p.id_product, p.product_name, p.id_parent_product
		FROM product p, nodes n
		WHERE p.id_parent_product = n.id_product
		AND p.price_description = ''
	)
	SELECT n.id_product, n.product_name FROM nodes n`

	rows, err := s.session.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vals := make([]bp.Category, 0, 32)
	for rows.Next() {
		var p bp.Category
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		vals = append(vals, p)
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

func (s Service) Products(searchQuery string) ([]bp.Product, error) {

	category := `2c7ea9a8-c9e1-4eee-ac31-d5e181bd8d06`
	searchQuery = strings.Replace(searchQuery, " ", "|", -1)
	searchQuery = strings.ToLower(searchQuery)

	query := `
	WITH RECURSIVE nodes (id_product, product_name, id_brand, weight, volume, id_parent_product, price_description, decimal_possibility)
	AS (
		SELECT p.id_product, p.product_name, p.id_brand, p.weight, p.volume, p.id_parent_product, p.price_description, p.decimal_possibility
		FROM product p
		WHERE p.id_parent_product = '{` + category + `}'
		UNION ALL

		SELECT p.id_product, p.product_name, p.id_brand, p.weight, p.volume, p.id_parent_product, p.price_description, p.decimal_possibility
		FROM product p, nodes n
		WHERE p.id_parent_product = n.id_product
	)
	SELECT * FROM nodes n
	WHERE lower(n.product_name) SIMILAR TO '%(` + searchQuery + `)%'
	AND NOT n.price_description=''`

	rows, err := s.session.db.Query(query)
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
