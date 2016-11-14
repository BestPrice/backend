package sql

import (
	"log"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/BestPrice/backend/bp"
)

var _ bp.Service = &Service{}

type Service struct {
	session *Session
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

	rows, err := s.session.db.Query(query)
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

	log.Println(phrase, " -> ", p)

	query := `
		WITH RECURSIVE nodes AS (
		-- GET all products with given category
		SELECT p.id_product uuid, p.price_description pd, ''::text || p.product_name AS chain
		FROM product p
		WHERE p.id_parent_product ` + c + `
		UNION ALL
		SELECT p.id_product, p.price_description, n.chain || ' ' || p.product_name
		FROM product p, nodes n
		WHERE p.id_parent_product = n.uuid
	), nodes2 AS (
		-- REMOVE category products and split chain
		SELECT n.uuid, regexp_split_to_table(n.chain, E'\\s+') words
		FROM nodes n
		WHERE NOT n.pd = ''
	), nodes3 AS (
		-- COUNT matches
		select n.uuid uuid, count(n.uuid) rank
		from nodes2 n
		WHERE unaccent(lower(n.words)) SIMILAR TO '%(` + p + `)%'
		group by n.uuid
	)
	SELECT p.*
	from product p, nodes3 n
	where n.uuid = p.id_product
	order by n.rank desc`

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
