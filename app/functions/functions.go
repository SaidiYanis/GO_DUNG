package functions

import (
	"encoding/json"
	"strings"

	"github.com/gofrs/uuid"
)

// NewUUID ...
func NewUUID() string {
	uuid, _ := uuid.NewV4()
	return uuid.String()
}

// RemoveDuplicate : Supression de doublons dans un tableau
func RemoveDuplicate(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		x = strings.TrimPrefix(x, "-")
		if !found[strings.ToLower(x)] {
			found[strings.ToLower(x)] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

// ConvertInputStructToDataStruct caution data must be an absolute pointer to work
func ConvertInputStructToDataStruct(input interface{}, data interface{}) error {
	// Sérialize Input in JSON
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return err
	}

	// Désérialize JSON in data
	err = json.Unmarshal(inputBytes, &data)
	if err != nil {
		return err
	}

	return nil
}
