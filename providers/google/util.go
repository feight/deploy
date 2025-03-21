package google

// Combine string slices
func env(e1 []string, e2 []string) []string {
	return append(e1, e2...)
}
