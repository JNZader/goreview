package git

// ParseDiff parses a unified diff string into a Diff struct.
// Full implementation in next commit.
func ParseDiff(diffText string) (*Diff, error) {
	return &Diff{
		Files: make([]FileDiff, 0),
	}, nil
}
