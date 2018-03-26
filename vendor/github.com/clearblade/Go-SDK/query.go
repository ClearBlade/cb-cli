package GoSDK

import ()

// Filter is the atomic structure inside a query it contains
// A field a value and an operator
type Filter struct {
	Field    string
	Value    interface{}
	Operator string
}

//Ordering dictates the order the query values are returned in. True is Ascending, False is Descending
type Ordering struct {
	SortOrder bool
	OrderKey  string
}

//Query contains configuration information for the request against a collection. It creates a subset of results for the operation to be performed upon
type Query struct {
	Filters    [][]Filter
	PageSize   int
	PageNumber int
	Order      []Ordering
	Columns    []string
}

//NewQuery allocates a new query
func NewQuery() *Query {
	query := &Query{
		Filters: [][]Filter{[]Filter{}},
		Order:   []Ordering{},
	}
	return query
}

//EqualTo adds an equality constraint to the query. Similar to "WHERE foo = 'bar'"
func (q *Query) EqualTo(field string, value interface{}) {
	f := Filter{
		Field:    field,
		Value:    value,
		Operator: "=",
	}
	q.Filters[0] = append(q.Filters[0], f)
}

//GreaterThan adds the corresponding constraint to the query. Similar to "WHERE foo > 3"
func (q *Query) GreaterThan(field string, value interface{}) {
	f := Filter{
		Field:    field,
		Value:    value,
		Operator: ">",
	}
	q.Filters[0] = append(q.Filters[0], f)
}

//GreaterThanEqualTo adds the corresponding constraint to the query. Similar to "WHERE foo >= 3"
func (q *Query) GreaterThanEqualTo(field string, value interface{}) {
	f := Filter{
		Field:    field,
		Value:    value,
		Operator: ">=",
	}
	q.Filters[0] = append(q.Filters[0], f)
}

//LessThan adds the corresponding constraint to the query. Similar to "WHERE foo < 3"
func (q *Query) LessThan(field string, value interface{}) {
	f := Filter{
		Field:    field,
		Value:    value,
		Operator: "<",
	}
	q.Filters[0] = append(q.Filters[0], f)
}

//LessThanEqualTo adds the corresponding constraint to the query. Similar to "WHERE foo <= 3"
func (q *Query) LessThanEqualTo(field string, value interface{}) {
	f := Filter{
		Field:    field,
		Value:    value,
		Operator: "<=",
	}
	q.Filters[0] = append(q.Filters[0], f)
}

//NotEqualTo adds the corresponding constraint to the query. Similar to "WHERE foo != 'bar'"
func (q *Query) NotEqualTo(field string, value interface{}) {
	f := Filter{
		Field:    field,
		Value:    value,
		Operator: "!=",
	}
	q.Filters[0] = append(q.Filters[0], f)
}

//Matches allows fuzzy matching on string columns. Use PCRE syntax.
func (q *Query) Matches(field, regex string) {
	f := Filter{
		Field:    field,
		Value:    regex,
		Operator: "~",
	}
	q.Filters[0] = append(q.Filters[0], f)
}

//Or applies an or constraint to the query.
func (q *Query) Or(orQuery *Query) {
	q.Filters = append(q.Filters, orQuery.Filters...)
}

// Map will produce the kind of thing that is sent as a query
// either as the body of a request or as a queryString
func (q *Query) serialize() map[string]interface{} {
	qrMap := make(map[string]interface{})
	qrMap["PAGENUM"] = q.PageNumber
	qrMap["PAGESIZE"] = q.PageSize
	qrMap["SELECTCOLUMNS"] = q.Columns
	sortMap := make([]map[string]interface{}, len(q.Order))
	for i, ordering := range q.Order {
		sortMap[i] = make(map[string]interface{})
		if ordering.SortOrder {
			sortMap[i]["ASC"] = ordering.OrderKey
		} else {
			sortMap[i]["DESC"] = ordering.OrderKey
		}
	}
	qrMap["SORT"] = sortMap
	filterSlice := make([][]map[string]interface{}, len(q.Filters))
	for i, querySlice := range q.Filters {
		qm := make([]map[string]interface{}, len(querySlice))
		for j, query := range querySlice {
			mapForQuery := make(map[string]interface{})
			var op string
			switch query.Operator {
			case "=":
				op = "EQ"
			case ">":
				op = "GT"
			case "<":
				op = "LT"
			case ">=":
				op = "GTE"
			case "<=":
				op = "LTE"
			case "/=", "!=":
				op = "NEQ"
			case "~":
				op = "RE"
			default:
				op = "WHAT_THE_HELL_ARE_YOU_DOING"
			}
			mapForQuery[op] = []map[string]interface{}{map[string]interface{}{query.Field: query.Value}}
			qm[j] = mapForQuery
		}
		filterSlice[i] = qm
	}
	qrMap["FILTERS"] = filterSlice
	return qrMap
}
