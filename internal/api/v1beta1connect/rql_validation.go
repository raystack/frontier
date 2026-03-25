package v1beta1connect

import "github.com/raystack/salt/rql"

// validateRQLQueryWithIlike validates an RQL query while allowing ilike/notilike operators.
// salt/rql does not include ilike/notilike in valid string operators, so we temporarily
// map them to like/notlike for validation only. The original operators are preserved
// in the query passed to the repository layer.
func validateRQLQueryWithIlike(q *rql.Query, schema any) error {
	vq := *q
	vq.Filters = make([]rql.Filter, len(q.Filters))
	for i, f := range q.Filters {
		nf := f
		switch nf.Operator {
		case "ilike":
			nf.Operator = "like"
		case "notilike":
			nf.Operator = "notlike"
		}
		vq.Filters[i] = nf
	}
	return rql.ValidateQuery(&vq, schema)
}
