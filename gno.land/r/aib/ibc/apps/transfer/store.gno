package transfer

import (
	"gno.land/p/nt/avl"
)

var denomsByBase *avl.Tree // ibcDenom:Denom

func init() {
	denomsByBase = avl.NewTree()
}

func hasDenomMetadata(voucherDenom string) bool {
	return denomsByBase.Has(voucherDenom)
}

func setDenomMetadata(denom Denom) {
	denomsByBase.Set(denom.IBCDenom(), denom)
}
