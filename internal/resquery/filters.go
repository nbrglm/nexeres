package resquery

import "slices"

type FilterMode string

const FilterModeAND FilterMode = "AND"
const FilterModeOR FilterMode = "OR"

type QueryFilterOp string

const (
	OpContains QueryFilterOp = "contains"
	OpEquals   QueryFilterOp = "equals"
	OpLte      QueryFilterOp = "lte"
	OpGte      QueryFilterOp = "gte"
	OpLt       QueryFilterOp = "lt"
	OpGt       QueryFilterOp = "gt"
)

type FilterOption struct {
	Field string        `json:"field"`
	Op    QueryFilterOp `json:"op"`
	Value any           `json:"value"`
}

// ValidateOption checks if the FilterOption's field and operation are valid
// based on the provided allowed fields and operations.
//
// Returns true if both the field and operation are valid; otherwise, returns false.
// If the FilterOption is nil, it is considered invalid.
func (o *FilterOption) ValidateOption(allowedFields []string, allowedOps []QueryFilterOp) bool {
	if o == nil {
		return false
	}
	fieldValid := slices.Contains(allowedFields, o.Field)
	opValid := slices.Contains(allowedOps, o.Op)

	return fieldValid && opValid
}

type QueryFilters struct {
	Filters []FilterOption `json:"options" binding:"required,min=1,dive,required"`
	Mode    FilterMode     `json:"mode" binding:"required"`
}

// Validate checks if all filters and the mode in QueryFilters are valid
// based on the provided allowed fields, operations, and modes.
//
// Returns true if all filters and the mode are valid; otherwise, returns false.
// If QueryFilters is nil, it is considered valid.
func (f *QueryFilters) Validate(allowedFields []string, allowedOps []QueryFilterOp, allowedModes []FilterMode) bool {
	if f == nil {
		return true
	}
	for _, filter := range f.Filters {
		if !filter.ValidateOption(allowedFields, allowedOps) {
			return false
		}
	}
	// All filters are valid, now check mode
	return slices.Contains(allowedModes, f.Mode)
}
