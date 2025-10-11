package matchpattern

func MatchPattern(patternParts, keyParts []string) bool {
	if len(patternParts) == 1 && patternParts[0] == "*" {
		return true
	}

	if len(patternParts) > len(keyParts) {
		return false
	}

	if len(patternParts) > 0 && patternParts[len(patternParts)-1] == "*" {
		if len(patternParts)-1 > len(keyParts) {
			return false
		}
		for i := 0; i < len(patternParts)-1; i++ {
			if patternParts[i] != "*" && patternParts[i] != keyParts[i] {
				return false
			}
		}
		return true
	}

	if len(patternParts) != len(keyParts) {
		return false
	}

	for i := 0; i < len(patternParts); i++ {
		if patternParts[i] != "*" && patternParts[i] != keyParts[i] {
			return false
		}
	}

	return true
}
