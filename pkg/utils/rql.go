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

type Group struct {
	Name string      `json:"name"`
	Data []GroupData `json:"data"`
}

type GroupData struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Page struct {
	// todo: update pkg/pagination/pagination.go to align with this
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
	TotalCount int64 `json:"total_count"`
}

func getFilterValueMethod(datatype string, filter *frontierv1beta1.RQLFilter) any {
	switch datatype {
	case "string", "datetime":
		return filter.GetStringValue()
	case "number":
		return filter.GetNumberValue()
	case "bool":
		return filter.GetBoolValue()
	default:
		return filter.GetStringValue()
	}
}

func TransformProtoToRQL(q *frontierv1beta1.RQLRequest, checkStruct interface{}) (*rql.Query, error) {
	filters := make([]rql.Filter, 0)
	for _, filter := range q.GetFilters() {
		datatype, err := rql.GetDataTypeOfField(filter.GetName(), checkStruct)
		if err != nil {
			return nil, err
		}
		filters = append(filters, rql.Filter{
			Name:     filter.GetName(),
			Operator: filter.GetOperator(),
			Value:    getFilterValueMethod(datatype, filter),
		})
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

func AddRQLPaginationInQuery(query *goqu.SelectDataset, rql *rql.Query) (*goqu.SelectDataset, Page) {
	// todo: update pkg/pagination/pagination.go to align with this
	var appliedLimit int
	var appliedOffset int

	if rql.Limit > 0 {
		appliedLimit = rql.Limit
	} else {
		appliedLimit = DefaultLimit
	}
	query = query.Limit(uint(appliedLimit))

	if rql.Offset > 0 {
		appliedOffset = rql.Offset
	} else {
		appliedOffset = DefaultOffset
	}
	query = query.Offset(uint(appliedOffset))

	return query, Page{
		Limit:  appliedLimit,
		Offset: appliedOffset,
	}
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

	selectCols := buildSelectColumns(groupByColumns)
	groupByCols := buildGroupByColumns(groupByColumns)

	query = query.
		Select(selectCols...).
		GroupBy(groupByCols...).
		Order(goqu.C(groupByColumns[0]).Asc())

	return query, nil
}

func buildGroupByColumns(columns []string) []interface{} {
	exprs := make([]interface{}, 0, len(columns))
	for _, col := range columns {
		exprs = append(exprs, goqu.L(col))
	}
	return exprs
}

func buildSelectColumns(columns []string) []interface{} {
	var valueExpr string
	switch len(columns) {
	case 1:
		valueExpr = fmt.Sprintf("%s AS values", columns[0])
	case 2:
		valueExpr = fmt.Sprintf("CONCAT(%s, ',', %s) AS values", columns[0], columns[1])
	default:
		valueExpr = fmt.Sprintf("%s AS values", columns[0]) // this function gets hit only for column length > 1,
		// so index 0 is always available.
	}

	return []interface{}{
		goqu.L(valueExpr),
		goqu.L("COUNT(*) as count"),
	}
}
