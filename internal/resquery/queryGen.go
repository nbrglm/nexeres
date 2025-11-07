package resquery

import (
	"fmt"
	"strings"
)

// Mapping of QueryFilterOp to PostgreSQL operators
var pgOps = map[QueryFilterOp]string{
	OpContains: "ILIKE",
	OpEquals:   "=",
	OpLte:      "<=",
	OpGte:      ">=",
	OpLt:       "<",
	OpGt:       ">",
}

// GeneratePGSQL generates a PostgreSQL query string based on the provided parameters.
// It constructs the WHERE clause using the specified filter mode (AND/OR) and applies pagination and sorting as needed.
//
// Parameters:
//
// - table: The name of the database table to query.
//
// - filterMode: The logical operator (AND/OR) to combine multiple filter conditions.
//
// - selectionFields: A list of fields to include in the SELECT statement. ADDS "COUNT(*) OVER() AS total_count" to the selection if fields are provided.
//
// - filters: A slice of FilterOption instances representing the filtering criteria.
//
// - pagination: An optional QueryPagination struct for limiting results and setting offsets.
//
// - sort: A slice of SortOption instances for ordering the results.
//
// Returns:
//
// - A SQL query string.
//
// - A slice of arguments corresponding to the filter values for parameterized queries.
//
// - An error if any issues occur during query generation.
func GeneratePGSQL(table string, filterMode FilterMode, selectionFields []string, filters []FilterOption, pagination *QueryPagination, sort []SortOption) (string, []any, error) {
	var whereClauses []string
	var args []any

	filterMode = FilterMode(strings.ToUpper(string(filterMode)))

	for i, filter := range filters {
		whereClauses = append(whereClauses, fmt.Sprintf("%s %s $%d", filter.Field, pgOps[filter.Op], i+1))
		val := filter.Value
		if filter.Op == OpContains {
			if s, ok := val.(string); ok {
				val = "%" + s + "%"
			}
		}
		args = append(args, val)
	}

	var query string

	if len(selectionFields) > 0 {
		selectionFields = append(selectionFields, "COUNT(*) OVER() AS total_count")
	} else {
		selectionFields = []string{"*", "COUNT(*) OVER() AS total_count"}
		// If no specific fields are selected, we select all fields and the total count
		// to ensure the total count is always available.
	}
	query = fmt.Sprintf("SELECT %s", strings.Join(selectionFields, ", "))
	query += fmt.Sprintf(" FROM %s", table)

	if len(whereClauses) > 0 {
		query += fmt.Sprintf(" WHERE %s", strings.Join(whereClauses, " "+string(filterMode)+" "))
	}

	if len(sort) > 0 {
		var sortClauses []string
		for _, s := range sort {
			sortClauses = append(sortClauses, fmt.Sprintf("%s %s", s.Field, strings.ToUpper(s.Direction)))
		}
		query += fmt.Sprintf(" ORDER BY %s", strings.Join(sortClauses, ", "))
	}

	if pagination == nil {
		pagination = DefaultQueryPagination()
	}

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pagination.PageSize, pagination.Page*pagination.PageSize)

	return query, args, nil
}
