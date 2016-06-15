package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/pborman/uuid"
)

//Generates a random WWN of the specified type:
//  - unit_serial: T10 WWN Unit Serial.
//  - iqn: iSCSI IQN
//  - naa: SAS NAA address
// @param wwn_type: The WWN address type.
// @type wwn_type: str
// @returns: A string containing the WWN.
func Generate_wwn(wwn_type string) (string, error) {
	switch strings.ToLower(wwn_type) {
	case "free":
		return uuid.New(), nil
	case "unit_serail":
		return uuid.New(), nil
	case "iqn":
		if localname, e := os.Hostname(); e == nil {
			localarch := strings.Replace(runtime.GOARCH, "_", "", -1)
			prefix := fmt.Sprintf("iqn.2016-06.org.linux-iscsi.%s.%s", strings.Split(localname, ".")[0], localarch)
			prefix = strings.ToLower(strings.TrimSpace(prefix))
			serial := fmt.Sprintf("sn.%s", uuid.New()[24:])
			return fmt.Sprintf("%s:%s", prefix, serial), nil
		} else {
			return "", e
		}
	case "naa":
		return "naa.5001405" + strings.Replace(uuid.New(), "-", "", -1)[23:], nil
	case "eui":
		return "eui.001405" + strings.Replace(uuid.New(), "-", "", -1)[22:], nil
	default:
		return "", fmt.Errorf("Unknown WWN type: %s", wwn_type)
	}
}
