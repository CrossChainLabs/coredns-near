// +build fuzz

package simple

import (
	"github.com/coredns/coredns/plugin/pkg/fuzz"
)

// Fuzz fuzzes cache.
func Fuzz(data []byte) int {
	w := Simple{}
	return fuzz.Do(w, data)
}
