package semver

import (
	"errors"
)

type ConstraintRangeEndpoint int

const (
	ConstraintRangeEndpointNil ConstraintRangeEndpoint = iota
	ConstraintRangeEndpointInclusive
	ConstraintRangeEndpointExclusive
)

type ConstraintRange struct {
	MatchPrerelease bool

	InvertRange bool
	Lower       ConstraintRangeEndpoint
	LowerValue  *Version
	Upper       ConstraintRangeEndpoint
	UpperValue  *Version
}

func (c *Constraints) AsRanges() ([][]ConstraintRange, error) {
	ret := make([][]ConstraintRange, 0, len(c.constraints))
	for _, or := range c.constraints {
		and := make([]ConstraintRange, 0, len(or))
		for _, expr := range or {
			e, err := constraintRangeOps[expr.origfunc](expr)
			if err != nil {
				return nil, err
			}
			and = append(and, e)
		}
		ret = append(ret, and)
	}
	return ret, nil
}

func EvalRanges(v *Version, r [][]ConstraintRange) bool {
orClause:
	for _, or := range r {
		for _, and := range or {
			if v.pre != "" && !and.MatchPrerelease {
				continue orClause
			}

			var matchRange = true
			switch and.Lower {
			case ConstraintRangeEndpointInclusive:
				matchRange = matchRange && v.Compare(and.LowerValue) >= 0
			case ConstraintRangeEndpointExclusive:
				matchRange = matchRange && v.Compare(and.LowerValue) > 0
			}
			switch and.Upper {
			case ConstraintRangeEndpointInclusive:
				matchRange = matchRange && v.Compare(and.UpperValue) <= 0
			case ConstraintRangeEndpointExclusive:
				matchRange = matchRange && v.Compare(and.UpperValue) < 0
			}
			if and.InvertRange {
				matchRange = !matchRange
			}
			if !matchRange {
				continue orClause
			}
		}
		return true
	}
	return false
}

var constraintRangeOps = map[string]func(*constraint) (ConstraintRange, error){
	"":   constraintRangeEqual,
	"=":  constraintRangeEqual,
	"!=": constraintRangeNotEqual,
	">":  constraintRangeGreaterThan,
	"<":  constraintRangeLessThan,
	">=": constraintRangeGreaterThanEqual,
	"=>": constraintRangeGreaterThanEqual,
	"<=": constraintRangeLessThanEqual,
	"=<": constraintRangeLessThanEqual,
	"~":  constraintRangeTilde,
	"~>": constraintRangeTilde,
	"^":  constraintRangeCaret,
}

func constraintRangeEqual(c *constraint) (ConstraintRange, error) {
	if c.dirty {
		return constraintRangeTilde(c)
	}

	return ConstraintRange{
		MatchPrerelease: c.con.pre != "",
		Lower:           ConstraintRangeEndpointInclusive,
		Upper:           ConstraintRangeEndpointInclusive,
		LowerValue:      c.con,
		UpperValue:      c.con,
	}, nil
}

func constraintRangeNotEqual(c *constraint) (ConstraintRange, error) {
	if !c.dirty {
		return ConstraintRange{
			MatchPrerelease: true,
			InvertRange:     true,
			Lower:           ConstraintRangeEndpointInclusive,
			Upper:           ConstraintRangeEndpointInclusive,
			LowerValue:      c.con,
			UpperValue:      c.con,
		}, nil
	}

	if c.con.pre != "" {
		return ConstraintRange{}, errors.New("unable to represent notequal constraint with prerelease")
	}

	var upper Version
	if c.minorDirty {
		upper = incKth(c.con, 0)
	} else if c.patchDirty {
		upper = incKth(c.con, 1)
	} else if c.extDirty {
		upper = incKth(c.con, len(c.con.ext)+2)
	} else {
		// major dirty
		return ConstraintRange{
			InvertRange: true,
			Lower:       ConstraintRangeEndpointInclusive,
			Upper:       ConstraintRangeEndpointInclusive,
			LowerValue:  c.con,
			UpperValue:  c.con,
		}, nil
	}

	return ConstraintRange{
		InvertRange: true,
		Lower:       ConstraintRangeEndpointInclusive,
		Upper:       ConstraintRangeEndpointExclusive,
		LowerValue:  c.con,
		UpperValue:  &upper,
	}, nil
}

//nolint:unparam
func constraintRangeGreaterThan(c *constraint) (ConstraintRange, error) {
	var lower Version
	if c.minorDirty {
		lower = incKth(c.con, 0)
	} else if c.patchDirty {
		lower = incKth(c.con, 1)
	} else if c.extDirty {
		lower = incKth(c.con, len(c.con.ext)+2)
	} else {
		// major dirty or not dirty
		return ConstraintRange{
			MatchPrerelease: c.con.pre != "",
			Lower:           ConstraintRangeEndpointExclusive,
			LowerValue:      c.con,
		}, nil
	}

	if c.con.pre != "" {
		lower.pre = "0"
	}
	return ConstraintRange{
		MatchPrerelease: c.con.pre != "",
		Lower:           ConstraintRangeEndpointInclusive,
		LowerValue:      &lower,
	}, nil
}

//nolint:unparam
func constraintRangeLessThan(c *constraint) (ConstraintRange, error) {
	return ConstraintRange{
		MatchPrerelease: c.con.pre != "",
		Upper:           ConstraintRangeEndpointExclusive,
		UpperValue:      c.con,
	}, nil
}

func constraintRangeGreaterThanEqual(c *constraint) (ConstraintRange, error) {
	cc, err := constraintRangeLessThan(c)
	cc.InvertRange = true
	return cc, err
}

func constraintRangeLessThanEqual(c *constraint) (ConstraintRange, error) {
	cc, err := constraintRangeGreaterThan(c)
	cc.InvertRange = true
	return cc, err
}

func constraintRangeTilde(c *constraint) (ConstraintRange, error) {
	if c.con.Major() == 0 && c.con.Minor() == 0 && c.con.Patch() == 0 && len(c.con.Ext()) == 0 &&
		!c.minorDirty && !c.patchDirty && !c.extDirty {
		if c.con.pre != "" {
			return ConstraintRange{
				MatchPrerelease: true,
				Lower:           ConstraintRangeEndpointInclusive,
				LowerValue:      c.con,
			}, nil
		}
		return ConstraintRange{}, nil
	}

	var upper Version
	if !c.dirty {
		// increase second to last
		upper = incKth(c.con, 2+len(c.con.ext)-1)
	} else if c.minorDirty {
		upper = incKth(c.con, 0)
	} else if c.patchDirty {
		upper = incKth(c.con, 1)
	} else if c.extDirty {
		upper = incKth(c.con, 2+len(c.con.ext))
	} else {
		panic("unreachable")
	}

	if c.con.pre != "" {
		upper.pre = "0"
	}
	return ConstraintRange{
		MatchPrerelease: c.con.pre != "",
		Lower:           ConstraintRangeEndpointInclusive,
		Upper:           ConstraintRangeEndpointExclusive,
		LowerValue:      c.con,
		UpperValue:      &upper,
	}, nil
}

//nolint:unparam
func constraintRangeCaret(c *constraint) (ConstraintRange, error) {
	var upper Version
	if c.con.Major() > 0 || c.minorDirty {
		upper = incKth(c.con, 0)
	} else if c.con.Minor() > 0 || c.patchDirty {
		upper = incKth(c.con, 1)
	} else if c.con.Patch() > 0 || len(c.con.ext) == 0 {
		upper = incKth(c.con, 2)
	} else {
		firstNoneZero := len(c.con.ext) - 1
		for i, e := range c.con.ext {
			if e != 0 {
				firstNoneZero = i
				break
			}
		}
		upper = incKth(c.con, 3+firstNoneZero)

	}

	if c.con.pre != "" {
		upper.pre = "0"
	}
	return ConstraintRange{
		MatchPrerelease: c.con.pre != "",
		Lower:           ConstraintRangeEndpointInclusive,
		Upper:           ConstraintRangeEndpointExclusive,
		LowerValue:      c.con,
		UpperValue:      &upper,
	}, nil
}

func incKth(v *Version, k int) Version {
	vNext := *v
	vNext.metadata = ""
	vNext.pre = ""
	if k == 0 {
		vNext.ext = nil
		vNext.patch = 0
		vNext.minor = 0
		vNext.major++
	} else if k == 1 {
		vNext.ext = nil
		vNext.patch = 0
		vNext.minor++
	} else if k == 2 {
		vNext.ext = nil
		vNext.patch++
	} else {
		vNext.ext = make([]uint64, k-2)
		for i := 0; i < k-3; i++ {
			vNext.ext[i] = v.ext[i]
		}
		vNext.ext[k-3] = v.ext[k-3] + 1
	}
	vNext.original = v.originalVPrefix() + "" + vNext.String()
	return vNext
}
