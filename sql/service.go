package sql

import (
	"database/sql"
	// "log"
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
		)	, join_brands AS (
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
	s.street_and_nr, s.district, s.region, s.latitude, s.longitude
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
		if err := rows.Scan(&s.ID, &s.CSName, &s.Name, &s.City, &s.Street, &s.District, &s.Region, &s.Lat, &s.Lng); err != nil {
			return nil, err
		}
		vals = append(vals, s)
	}

	return vals, nil
}

func (s *Service) shopQuery(ID bp.ID) string {
	id := ID.String()
	query := `
WITH RECURSIVE
t0 AS (
	SELECT p.id_product
	FROM product p
	WHERE p.id_parent_product = '` + id + `'
	UNION ALL
	SELECT p.id_product
	FROM product p, t0 n
	WHERE p.id_parent_product = n.id_product
)
, t1 AS (
	SELECT t.id_product FROM t0 t
	UNION
	SELECT '` + id + `'
)
, t2 AS (
	SELECT pp.*
	FROM t1 t, product_prices pp
	WHERE t.id_product = pp.id_product
)
, t3 AS (
	SELECT '` + id + `' as id_product, cs.chain_store_name, p.product_name, b.brand_name, p.price_description, t.unit_price,
	cs.id_chain_store
	--, p.weight, p.volume, p.decimal_possibility
	FROM t2 t
	JOIN product p ON p.id_product = t.id_product
	JOIN chain_store cs ON cs.id_chain_store = t.id_chain_store
	JOIN brand b ON b.id_brand = p.id_brand
)
SELECT * FROM t3
`
	return query
}

func (s Service) Shop(req *bp.ShopRequest) (bp.Shop, error) {
	var (
		IDs []string
		p   []bp.ShopProduct
	)
	for _, product := range req.Products {
		IDs = append(IDs, product.ID.String())

		rows, err := s.db.Query(s.shopQuery(product.ID))
		if err != nil {
			return bp.Shop{}, err
		}
		defer rows.Close()

		for rows.Next() {
			var r bp.ShopProduct
			err := rows.Scan(&r.ID, &r.ChainStore, &r.Product,
				&r.Brand, &r.PriceDesc, &r.Price, &r.IDChainStore)
			if err != nil {
				return bp.Shop{}, err
			}
			p = append(p, r)
		}
	}

	return calcShop(p, req)
}

type Stores []bp.ShopStore

type shopProducts struct {
	p []bp.ShopProduct
}

func (b *shopProducts) Len() int      { return len(b.p) }
func (b *shopProducts) Swap(i, j int) { b.p[i], b.p[j] = b.p[j], b.p[i] }

type byPrice struct {
	shopProducts
}

func (b *byPrice) Less(i, j int) bool {
	return b.p[i].Price.Cmp(b.p[j].Price) < 0
}

func calcShop(p []bp.ShopProduct, req *bp.ShopRequest) (bp.Shop, error) {

	var (
		stores     = make(map[string]*bp.ShopStore)
		m          = make(map[string]bool)
		priceTotal decimal.Decimal
	)

	// add price to products
	for i := range p {
		pid := p[i].ID.String()
		m[pid] = false
		p[i].Count = req.ProductCount(p[i].ID)
		p[i].Price = p[i].Price.Mul(decimal.NewFromFloat(float64(p[i].Count)))
	}

	sort.Sort(&byPrice{shopProducts{p}})
	findProducts(p, req, stores, make(map[string]bool))

	var (
		Stores        Stores
		productsTotal int
	)
	for _, store := range stores {
		// remove not prefered chainstores
		if !req.UserPreference.Contains(store.ID) {
			continue
		}
		for _, product := range store.Products {
			priceTotal = priceTotal.Add(product.Price)
		}
		productsTotal += len(store.Products)
		Stores = append(Stores, *store)
	}

	if productsTotal != len(req.Products) {
		return bp.Shop{Error: "one or more products not available in store"}, nil
	}

	return bp.Shop{
		Stores:     Stores,
		PriceTotal: priceTotal,
	}, nil
}

func findProducts(products []bp.ShopProduct, req *bp.ShopRequest, stores map[string]*bp.ShopStore, pt map[string]bool) {

	for i, p := range products {
		pid := p.ID.String()
		if _, ok := pt[pid]; ok {
			continue
		}
		pt[pid] = true

		if len(pt) > len(req.Products) {
			delete(pt, pid)
			return
		}

		pidcs := p.IDChainStore.String()
		store := stores[pidcs]
		if store == nil {
			store = &bp.ShopStore{
				ID:             p.IDChainStore,
				ChainStoreName: p.ChainStore,
			}
			stores[pidcs] = store
		}

		if len(stores) > req.UserPreference.MaxStores {
			delete(stores, pidcs)
			delete(pt, pid)
			continue
		}

		store.Products = append(store.Products, p)

		findProducts(products[i+1:], req, stores, pt)

		if len(pt) == len(req.Products) {
			return
		}

		store.Products = store.Products[:len(store.Products)-1]
		if len(store.Products) == 0 {
			delete(stores, pidcs)
		}

		delete(pt, pid)
	}
}
