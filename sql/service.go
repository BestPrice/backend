package sql

import (
	"database/sql"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/BestPrice/backend/bp"
)

var _ bp.Service = &Service{}

type Service struct {
	db *sql.DB
}

func makeCategoryTree(parent *bp.ID, cat map[*bp.Category]bool) []bp.Category {
	nodes := []bp.Category{}

	for c := range cat {
		if parent == nil {
			if !c.IDParent.Null() {
				continue
			}
		} else {
			if c.IDParent.Null() || c.IDParent.String() != parent.String() {
				continue
			}
		}
		delete(cat, c)
		c.Subcategories = makeCategoryTree(&c.ID, cat)
		nodes = append(nodes, *c)
	}

	return nodes
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
	SELECT n.id_product, n.product_name, n.id_parent_product FROM nodes n`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vals := make(map[*bp.Category]bool)
	for rows.Next() {
		var p bp.Category
		if err := rows.Scan(&p.ID, &p.Name, &p.IDParent); err != nil {
			return nil, err
		}
		vals[&p] = true
	}

	return makeCategoryTree(nil, vals), nil
}

func (s Service) Chainstores() ([]bp.Chainstore, error) {
	rows, err := s.db.Query("SELECT * FROM chain_store")
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

func normalizePhrase(p string) (string, error) {
	var (
		add = func(r rune) rune {
			if r == ' ' {
				return '|'
			}
			return r
		}
	)

	p = strings.Replace(p, "|", "", -1)
	p = strings.TrimSpace(p)

	t := transform.Chain(
		runes.Map(add),
		runes.Map(unicode.ToLower),
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC)
	no, _, err := transform.String(t, p)
	return no, err
}

func (s Service) Products(category *bp.ID, phrase string) ([]bp.Product, error) {

	c := "IS NULL"

	if category != nil {
		c = "= '{" + category.String() + "}'"
	}

	p, err := normalizePhrase(phrase)
	if err != nil {
		return nil, err
	}

	query := `
		WITH RECURSIVE nodes AS (
			-- GET all products with given category
			SELECT p.id_product uuid, p.id_brand, p.price_description pd, ''::text || p.product_name AS chain
			FROM product p
			WHERE p.id_parent_product ` + c + `
			UNION ALL
			SELECT p.id_product, p.id_brand, p.price_description, n.chain || ' ' || p.product_name
			FROM product p, nodes n
			WHERE p.id_parent_product = n.uuid
		), join_brands AS (
			SELECT n.uuid, n.pd, n.chain || ' ' || b.brand_name AS chain
			FROM nodes n
			JOIN brand b ON b.id_brand = n.id_brand
		), nodes2 AS (
			-- REMOVE category products and split chain
			SELECT n.uuid, regexp_split_to_table(n.chain, E'\\s+') words
			FROM join_brands n
			WHERE NOT n.pd = ''
		), nodes3 AS (
			-- COUNT matches
			SELECT n.uuid uuid, count(n.uuid) rank
			FROM nodes2 n
			WHERE unaccent(lower(n.words)) SIMILAR TO '%(` + p + `)%'
			GROUP BY n.uuid
		), nodes4 AS (
			SELECT p.*, n.rank
			FROM product p, nodes3 n
			WHERE p.id_product = n.uuid
		)
		SELECT
		n.id_product,
		n.product_name,
		n.weight,
		n.volume,
		n.price_description,
		n.decimal_possibility,
		b.id_brand,
		b.brand_name,
		n.rank
		FROM nodes4 n
		JOIN brand b ON b.id_brand = n.id_brand
		ORDER BY n.rank DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vals := make([]bp.Product, 0, 32)
	for rows.Next() {
		var p bp.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Weight, &p.Volume, &p.PriceDescription,
			&p.DecimalPossibility, &p.Brand.ID, &p.Brand.Name, &p.Rank); err != nil {
			return nil, err
		}
		vals = append(vals, p)
	}

	return vals, nil
}

func (s Service) Stores(chainstore, district, region string) ([]bp.Store, error) {
	query := `
	SELECT s.id_store, cs.chain_store_name, s.store_name, s.city,
	s.street_and_nr, s.district, s.region, s.coordinates
	FROM store s
	JOIN chain_store cs ON s.id_chain_store = cs.id_chain_store
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vals := make([]bp.Store, 0, 32)
	for rows.Next() {
		var s bp.Store
		if err := rows.Scan(&s.ID, &s.CSName, &s.Name, &s.City, &s.Street, &s.District, &s.Region, &s.Coordinates); err != nil {
			return nil, err
		}
		vals = append(vals, s)
	}

	return vals, nil
}

func (s Service) Shop() (bp.Shop, error) {
	panic("TODO: implement Shop")
}
