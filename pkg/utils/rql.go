package utils

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"golang.org/x/exp/slices"
)

const (
	OperatorEmpty    = "empty"
	OperatorNotEmpty = "notempty"
	OperatorIn       = "in"
	OperatorNotIn    = "notin"
	OperatorLike     = "like"
	OperatorNotLike  = "notlike"
	DefaultLimit     = 50
	DefaultOffset    = 0
)

func TransformProtoToRQL(q *frontierv1beta1.RQLRequest, checkStruct interface{}) (*rql.Query, error) {
	filters := make([]rql.Filter, 0)
	for _, filter := range q.GetFilters() {
		datatype, err := rql.GetDataTypeOfField(filter.GetName(), checkStruct)
		if err != nil {
			return nil, err
		}
		switch datatype {
		case "string":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetStringValue(),
			})
		case "number":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetNumberValue(),
			})
		case "bool":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetBoolValue(),
			})
		case "datetime":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetStringValue(),
			})
		}
	}

	sortItems := make([]rql.Sort, 0)
	for _, sortItem := range q.GetSort() {
		sortItems = append(sortItems, rql.Sort{Name: sortItem.GetName(), Order: sortItem.GetOrder()})
	}

	return &rql.Query{
		Search:  q.GetSearch(),
		Offset:  int(q.GetOffset()),
		Limit:   int(q.GetLimit()),
		Filters: filters,
		Sort:    sortItems,
		GroupBy: q.GetGroupBy(),
	}, nil
}

func AddRQLSortInQuery(query *goqu.SelectDataset, rql *rql.Query) (*goqu.SelectDataset, error) {
	// If there is a group by parameter added then sort the result
	// by group_by first key in asc order by default before any other sort column

	for _, groupBy := range rql.GroupBy {
		query = query.OrderAppend(goqu.C(groupBy).Asc())
	}

	for _, sortItem := range rql.Sort {
		switch sortItem.Order {
		case "asc":
			query = query.OrderAppend(goqu.C(sortItem.Name).Asc())
		case "desc":
			query = query.OrderAppend(goqu.C(sortItem.Name).Desc())
		default:
			query = query.OrderAppend(goqu.C(sortItem.Name).Asc())
		}
	}
	return query, nil
}

func AddRQLSearchInQuery(query *goqu.SelectDataset, rql *rql.Query, rqlSearchSupportedColumns []string) (*goqu.SelectDataset, error) {
	// this should contain only those columns that are sql string(text, varchar etc) datatype

	searchExpressions := make([]goqu.Expression, 0)
	if rql.Search != "" {
		for _, col := range rqlSearchSupportedColumns {
			searchExpressions = append(searchExpressions, goqu.L(
				fmt.Sprintf(`"%s"::TEXT ILIKE '%%%s%%'`, col, rql.Search),
			))
		}
	}
	return query.Where(goqu.Or(searchExpressions...)), nil
}

func AddRQLFiltersInQuery(query *goqu.SelectDataset, rqlInput *rql.Query, rqlFilerSupportedColumns []string, checkStruct interface{}) (*goqu.SelectDataset, error) {
	for _, filter := range rqlInput.Filters {
		if !slices.Contains(rqlFilerSupportedColumns, filter.Name) {
			return nil, fmt.Errorf("%s is not supported in filters", filter.Name)
		}
		datatype, err := rql.GetDataTypeOfField(filter.Name, checkStruct)
		if err != nil {
			return query, err
		}
		switch datatype {
		case "string":
			query = ProcessStringDataType(filter, query)
		case "number":
			query = query.Where(goqu.Ex{
				filter.Name: goqu.Op{filter.Operator: filter.Value.(float32)},
			})
		case "bool":
			query = query.Where(goqu.Ex{
				filter.Name: goqu.Op{filter.Operator: filter.Value.(bool)},
			})
		case "datetime":
			query = query.Where(goqu.Ex{
				filter.Name: goqu.Op{filter.Operator: filter.Value.(string)},
			})
		}
	}
	return query, nil
}

func ProcessStringDataType(filter rql.Filter, query *goqu.SelectDataset) *goqu.SelectDataset {
	switch filter.Operator {
	case OperatorEmpty:
		query = query.Where(goqu.L(fmt.Sprintf("coalesce(%s, '') = ''", filter.Name)))
	case OperatorNotEmpty:
		query = query.Where(goqu.L(fmt.Sprintf("coalesce(%s, '') != ''", filter.Name)))
	case OperatorIn, OperatorNotIn:
		// process the values of in and notin operators as comma separated list
		fmt.Println("filter.Value.(string) for in and not in", filter.Value.(string))
		query = query.Where(goqu.Ex{
			filter.Name: goqu.Op{filter.Operator: strings.Split(filter.Value.(string), ",")},
		})
	case OperatorLike:
		// some semi-string sql types like UUID require casting to text to support like operator
		query = query.Where(goqu.L(fmt.Sprintf(`"%s"::TEXT ILIKE '%s'`, filter.Name, filter.Value.(string))))
	case OperatorNotLike:
		// some semi-string sql types like UUID require casting to text to support like operator
		query = query.Where(goqu.L(fmt.Sprintf(`"%s"::TEXT NOT ILIKE '%s'`, filter.Name, filter.Value.(string))))
	default:
		query = query.Where(goqu.Ex{filter.Name: goqu.Op{filter.Operator: filter.Value.(string)}})
	}
	return query
}

func AddRQLPaginationInQuery(query *goqu.SelectDataset, rql *rql.Query) *goqu.SelectDataset {
	// todo: update pkg/pagination/pagination.go to align with this
	if rql.Limit > 0 {
		query = query.Limit(uint(rql.Limit))
	} else {
		query = query.Limit(DefaultLimit)
	}
	if rql.Offset > 0 {
		query = query.Offset(uint(rql.Offset))
	} else {
		query = query.Offset(DefaultOffset)
	}
	return query
}

func AddGroupInQuery(query *goqu.SelectDataset, rql *rql.Query, allowedGroupByColumns []string) (*goqu.SelectDataset, error) {
	groupByColumns := rql.GroupBy
	if len(groupByColumns) == 0 {
		return query, nil
	}

	for _, col := range groupByColumns {
		if !slices.Contains(allowedGroupByColumns, col) {
			return nil, fmt.Errorf("%s is not supported in group by", col)
		}
	}

	query = query.
		Select(buildGroupByColumns(groupByColumns)...).
		GroupBy(buildGroupByColumns(groupByColumns)[:len(groupByColumns)]...).
		Order(goqu.C(groupByColumns[0]).Asc())

	return query, nil
}

func buildGroupByColumns(columns []string) []interface{} {
	exprs := make([]interface{}, 0, len(columns)+1)
	for _, col := range columns {
		exprs = append(exprs, goqu.L(col))
	}
	return append(exprs, goqu.COUNT("*"))
}
