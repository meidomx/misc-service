package ldap

import "strings"

func SplitDN(dn string) []string {
	ss := strings.Split(dn, ",")
	rs := make([]string, len(ss))
	for i, k := range ss {
		rs[i] = strings.TrimSpace(k)
	}
	return rs
}

func CombineParentDN(levels []string) string {
	if len(levels) <= 1 {
		return ""
	} else {
		return strings.Join(levels[1:], ",")
	}
}

func CombineDN(levels []string) string {
	return strings.Join(levels, ",")
}

func EntryType(dni string) string {
	return strings.Split(dni, "=")[0]
}
