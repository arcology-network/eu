package shared

import "encoding/gob"

func init() {
	gob.Register(&[]*EuResult{})
	gob.Register(&Euresults{})
}
