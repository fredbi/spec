package spec

import "net/url"

func isRefLocal(ref Ref) bool {
	if ref.IsRoot() {
		return true
	}
	// this supports formally invalid but tolerated, URI forms: C:\x\y\z (strict URI should be file://C:\x\y\z)
	u, _ := url.Parse(ref.String())
	return ref.HasFragmentOnly && u.Opaque == ""
}
