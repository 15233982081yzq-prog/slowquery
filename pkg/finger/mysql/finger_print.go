package mysql

import (
	tidbParser "github.com/pingcap/tidb/parser"
)

type FingerPrint struct {
	keepHint bool
}

func NewFingerPrint(keepHint bool) *FingerPrint {
	return &FingerPrint{
		keepHint: keepHint,
	}
}

func (fp *FingerPrint) NormalizeDigest(sql string) (string, *tidbParser.Digest) {
	if fp.keepHint {
		normalized := tidbParser.NormalizeKeepHint(sql)
		return normalized, tidbParser.DigestNormalized(normalized)
	}

	return tidbParser.NormalizeDigest(sql)
}

func (fp *FingerPrint) Normalize(sql string) string {
	if fp.keepHint {
		return tidbParser.NormalizeKeepHint(sql)
	}
	return tidbParser.Normalize(sql)
}

func (fp *FingerPrint) DigestNormalize(normalize string) *tidbParser.Digest {
	return tidbParser.DigestNormalized(normalize)
}

func (fp *FingerPrint) Digest(sql string) *tidbParser.Digest {
	return tidbParser.DigestNormalized(tidbParser.Normalize(sql))
}
