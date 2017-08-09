package soapboxd

import "database/sql"

func newNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
