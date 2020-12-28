package grab

import (
	"fmt"
)

// FlagMap is a function that maps a single-bit bitmask (i.e. a number of the
// form (1 << x)) to a string representing that bit.
// If the input is not valid / recognized, it should return a non-nil error,
// which will cause the flag to be added to the "unknowns" list.
type FlagMap func(uint64) (string, error)

// MapFlagsToSet gets the "set" (map of strings to true) of values corresponding
// to the bits in flags. For each bit i set in flags, the result will have
// result[mapping(i << i)] = true.
// Any bits for which the mapping returns a non-nil error are instead appended
// to the unknowns list.
func MapFlagsToSet(flags uint64, mapping FlagMap) (map[string]bool, []uint64) {
	ret := make(map[string]bool)
	unknowns := []uint64{}
	for i := uint8(0); i < 64; i++ {
		if flags == 0 {
			break
		}
		bit := (flags & 1) << i
		if bit > 0 {
			str, err := mapping(bit)
			if err != nil {
				unknowns = append(unknowns, bit)
			} else {
				ret[str] = true
			}
		}
		flags >>= 1
	}
	return ret, unknowns
}

// GetFlagMapFromMap returns a FlagMap function that uses mapping to do the
// mapping. Values not present in the map are treated as unknown, and a non-nil
// error is returned in those cases.
func GetFlagMapFromMap(mapping map[uint64]string) FlagMap {
	return func(bit uint64) (string, error) {
		ret, ok := mapping[bit]
		if ok {
			return ret, nil
		}
		return "", fmt.Errorf("Unknown flag 0x%x", bit)
	}
}

// GetFlagMapFromList returns a FlagMap function mapping the ith bit to the
// ith entry of bits.
// bits is a list of labels for the corresponding bits; any empty strings (and
// bits beyond the end of the list) are treated as unknown.
func GetFlagMapFromList(bits []string) FlagMap {
	mapping := make(map[uint64]string)
	for i, v := range bits {
		if v != "" {
			mapping[uint64(1)<<uint8(i)] = v
		}
	}
	return GetFlagMapFromMap(mapping)
}

// ListFlagsToSet converts an integer flags variable to a set of string labels
// corresponding to each bit, in the format described by the wiki (see
// https://github.com/zmap/zgrab2/wiki/Scanner-details).
// The ith entry of labels gives the label for the ith bit (i.e. flags & (1<<i)).
// Empty strings in labels are treated as unknown, as are bits beyond the end
// of the list. Unknown flags are appended to the unknown list.
func ListFlagsToSet(flags uint64, labels []string) (map[string]bool, []uint64) {
	mapper := GetFlagMapFromList(labels)
	return MapFlagsToSet(flags, mapper)
}
