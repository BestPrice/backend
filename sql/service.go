package sql

import (
	"database/sql"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/BestPrice/backend/bp"
	"github.com/shopspring/decimal"
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

func (s Service) Stores() ([]bp.Store, error) {
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

func (s Service) Shop(req *bp.ShopRequest) (bp.Shop, error) {
	var IDs []string
	for _, p := range req.Products {
		IDs = append(IDs, `"`+p.ID.String()+`"`)
	}

	query := `
WITH 
t0 AS (
	SELECT * FROM product_prices pp
	WHERE pp.id_product = ANY('{` + strings.Join(IDs, ",") + `}'::uuid[])
)
, t1 AS (
	SELECT p.id_product, cs.chain_store_name, p.product_name, b.brand_name, p.price_description, t.unit_price,
	cs.id_chain_store
	--, p.weight, p.volume, p.decimal_possibility
	FROM t0 t
	JOIN product p ON p.id_product = t.id_product
	JOIN chain_store cs ON cs.id_chain_store = t.id_chain_store
	JOIN brand b ON b.id_brand = p.id_brand
)
SELECT * FROM t1
`
	rows, err := s.db.Query(query)
	if err != nil {
		return bp.Shop{}, err
	}
	defer rows.Close()

	var p []bp.ShopProduct
	for rows.Next() {
		var r bp.ShopProduct
		err := rows.Scan(&r.ID, &r.ChainStore, &r.Product, &r.Brand, &r.PriceDesc, &r.Price,
			&r.IDChainStore)
		if err != nil {
			return bp.Shop{}, err
		}
		p = append(p, r)
	}

	return calcShop(p, req)
}

type Stores []bp.ShopStore

type byPrice struct {
	p []bp.ShopProduct
}

func (b *byPrice) Len() int           { return len(b.p) }
func (b *byPrice) Less(i, j int) bool { return b.p[i].Price.Cmp(b.p[j].Price) < 0 }
func (b *byPrice) Swap(i, j int)      { b.p[i], b.p[j] = b.p[j], b.p[i] }

func calcShop(p []bp.ShopProduct, req *bp.ShopRequest) (bp.Shop, error) {

	var (
		stores      = make(map[string]bp.ShopStore)
		m           = make(map[string]bool)
		priceTotal  decimal.Decimal
		preferedSet = false
	)

	// add price to products
	for i := range p {
		m[p[i].ID.String()] = false
		p[i].Count = req.ProductCount(p[i].ID)
		p[i].Price = p[i].Price.Mul(decimal.NewFromFloat(float64(p[i].Count)))
	}

	// sort by price
	sort.Sort(&byPrice{p})
	for _, pr := range p {
		if !req.UserPreference.Contains(pr.IDChainStore) || m[pr.ID.String()] {
			continue
		}
		if len(stores) == req.UserPreference.MaxStores && !preferedSet {
			var prefered []bp.ID
			for _, s := range stores {
				prefered = append(prefered, s.Products[0].IDChainStore)
			}
			req.UserPreference.SetPrefered(prefered)

			preferedSet = true
			continue
		}

		m[pr.ID.String()] = true

		key := pr.IDChainStore.String()
		s := stores[key]

		s.ChainStoreName = pr.ChainStore
		priceTotal = priceTotal.Add(pr.Price)

		s.Products = append(s.Products, pr)
		stores[key] = s
	}

	if len(stores) == 0 {
		return bp.Shop{Error: "specified products not found in specified chainstores"}, nil
	}

	// remove chainstores which does not have all products
	var s Stores
	for _, v := range stores {
		s = append(s, v)
	}

	if len(s) == 0 {
		return bp.Shop{Error: "one or more products not available in store"}, nil
	}

	return bp.Shop{
		Stores:     s,
		PriceTotal: priceTotal,
	}, nil
}
