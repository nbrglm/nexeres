package resquery

import "slices"

type SortOption struct {
	Field     string `json:"field" binding:"required"`
	Direction string `json:"direction" binding:"required,oneof=asc desc ASC DESC"`
}

// ValidateOption checks if the SortOption's field is valid
// based on the provided allowed fields.
//
// Returns true if the field is valid; otherwise, returns false.
// If the SortOption is nil, it is considered invalid.
func (o *SortOption) ValidateOption(allowedFields []string) bool {
	if o == nil {
		return false
	}
	fieldValid := slices.Contains(allowedFields, o.Field)
	return fieldValid
}

func ValidateSortOptions(options []SortOption, allowedFields []string) bool {
	for _, option := range options {
		if !option.ValidateOption(allowedFields) {
			return false
		}
	}
	return true
}
