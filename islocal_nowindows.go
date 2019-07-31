// +build !windows

package spec

func isRefLocal(ref Ref) bool {
	return (ref.IsRoot() || ref.HasFragmentOnly)
}
