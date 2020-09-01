package modem

import "testing"

func TestRegExp(t *testing.T) {
	var str = `var tagValueList = '531000000|Locked|OK|Operational|OK|Operational|&nbsp;|&nbsp;|Enabled|BPI+|Wed Aug 26 21:43:03 2020|0|0|0|04:59:53|3|';`
	b := CM1200VarRegExp.Match([]byte(str))
	if !b {
		t.Fail()
	}
}
