package dbutil

import (
	"fmt"
	"net/url"
)

// SetDBSchema add dbschema to the dbsource's url if search_path not defined.
func SetDBSchema(dbSource, dbSchema string) (string, error) {
	if dbSource == "" || dbSchema == "" {
		return dbSource, nil
	}

	u, err := url.Parse(dbSource)
	if err != nil {
		return dbSource, fmt.Errorf("failed to parse url")
	}

	if u.Query().Get("search_path") == "" {
		qValues := u.Query()
		qValues.Set("search_path", dbSchema)

		u.RawQuery = qValues.Encode()

		return u.String(), nil
	}

	return dbSource, nil
}

// GetDBSchema get dbschema from the dbsource's url.
//   - return empty string if search_path not defined and parse failed.
func GetDBSchema(dbSource string) string {
	u, err := url.Parse(dbSource)
	if err != nil {
		return ""
	}

	return u.Query().Get("search_path")
}
