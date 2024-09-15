package tenders

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
)

func ParseLimitOffset(rawQuery string) (int, int, error) {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return 0, 0, err
	}
	limit, offset := 0, 0
	limitArgs, ok := values["limit"]
	if ok {
		res, err := strconv.Atoi(limitArgs[len(limitArgs)-1])
		if err != nil {
			return 0, 0, err
		}
		if res < 0 {
			return 0, 0, fmt.Errorf("invalid value for limit (%v)", res)
		}
		limit = res
	} else {
		limit = 5
	}
	offsetArgs, ok := values["offset"]
	if ok {
		res, err := strconv.Atoi(offsetArgs[len(offsetArgs)-1])
		if err != nil {
			return 0, 0, err
		}
		if res < 0 {
			return 0, 0, fmt.Errorf("invalid value for offset (%v)", res)
		}
		offset = res
	}
	return limit, offset, nil
}

var statusType = []string{
	"Construction",
	"Delivery",
	"Manufacture",
}

func ParseLimitOffsetService(rawQuery string) (int, int, []string, error) {
	limit, offset, err := ParseLimitOffset(rawQuery)
	if err != nil {
		return 0, 0, nil, err
	}
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return 0, 0, nil, err
	}
	services, ok := values["service_type"]
	if !ok {
		return limit, offset, statusType, nil
	}
	for _, service := range services {
		if !slices.Contains(statusType, service) {
			return 0, 0, nil, fmt.Errorf("invalid service_type: %v", service)
		}
	}
	return limit, offset, services, nil
}
