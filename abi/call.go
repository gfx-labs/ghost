package abi

func EncodeCallData(sig Signature, dat func(*Builder) *Builder) []byte {
	b := new(Builder)
	b = dat(b)
	ans := append(sig.SelectorB(), b.Finish()...)
	return ans
}
