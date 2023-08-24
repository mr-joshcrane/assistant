package assistant

import (
	"github.com/mr-joshcrane/oracle"
)

func Start() error {
	o := oracle.NewOracle()
	o.SetPurpose("You are a helpful assistant.")

	return nil
}
