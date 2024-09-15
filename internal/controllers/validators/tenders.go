package validators

import (
	"fmt"
)

func EditTenderValidate(name, description *string) error {
	if name != nil {
		if *name == "" {
			return fmt.Errorf("invalid params: name")
		}
	}
	if description != nil {
		if *description == "" {
			return fmt.Errorf("invalid params: description")
		}
	}
	return nil
}
