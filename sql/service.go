package sql

import "github.com/BestPrice/backend/bp"

var _ bp.Service = &Service{}

type Service struct {
	session *Session
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
		if err := rows.Scan(&c.Id, &c.Name); err != nil {
			return nil, err
		}
		vals = append(vals, c)
	}

	return vals, nil
}

func (s Service) Categories() ([]bp.Category, error) {
	rows, err := s.session.db.Query("SELECT * FROM category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vals := make([]bp.Category, 0, 32)
	for rows.Next() {
		var c bp.Category
		if err := rows.Scan(&c.Id, &c.Name, &c.IdParent); err != nil {
			return nil, err
		}
		vals = append(vals, c)
	}

	// TODO: make category "tree"
	//
	// sort.Sort(&categoriesById{vals})
	// var categories []bp.Category

	// for len(vals) > 0 {

	// }

	return vals, nil
}

func (s Service) Products(query string) ([]bp.Product, error) {
	rows, err := s.session.db.Query("SELECT * FROM product LIMIT 100")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vals := make([]bp.Product, 0, 32)
	for rows.Next() {
		var p bp.Product
		if err := rows.Scan(&p.Id); err != nil {
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
		if err := rows.Scan(&s.Id); err != nil {
			return nil, err
		}
		vals = append(vals, s)
	}

	return vals, nil
}

func (s Service) Shop() (bp.Shop, error) {
	panic("TODO: implement Shop")
}

type categoriesById struct {
	c []bp.Category
}

func (c *categoriesById) Len() int           { return len(c.c) }
func (c *categoriesById) Less(i, j int) bool { return c.c[i].Id < c.c[j].Id }
func (c *categoriesById) Swap(i, j int)      { c.c[i], c.c[j] = c.c[j], c.c[i] }
