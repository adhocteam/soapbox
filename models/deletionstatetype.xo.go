// Package models contains the types for schema 'public'.
package models

// GENERATED BY XO. DO NOT EDIT.

import (
	"database/sql/driver"
	"errors"
)

// DeletionStateType is the 'deletion_state_type' enum type from schema 'public'.
type DeletionStateType uint16

const (
	// DeletionStateTypeNotDeleted is the 'NOT_DELETED' DeletionStateType.
	DeletionStateTypeNotDeleted = DeletionStateType(1)

	// DeletionStateTypeDeleteInfrastructureWait is the 'DELETE_INFRASTRUCTURE_WAIT' DeletionStateType.
	DeletionStateTypeDeleteInfrastructureWait = DeletionStateType(2)

	// DeletionStateTypeDeleteInfrastructureSucceeded is the 'DELETE_INFRASTRUCTURE_SUCCEEDED' DeletionStateType.
	DeletionStateTypeDeleteInfrastructureSucceeded = DeletionStateType(3)

	// DeletionStateTypeDeleteInfrastructureFailed is the 'DELETE_INFRASTRUCTURE_FAILED' DeletionStateType.
	DeletionStateTypeDeleteInfrastructureFailed = DeletionStateType(4)
)

// String returns the string value of the DeletionStateType.
func (dst DeletionStateType) String() string {
	var enumVal string

	switch dst {
	case DeletionStateTypeNotDeleted:
		enumVal = "NOT_DELETED"

	case DeletionStateTypeDeleteInfrastructureWait:
		enumVal = "DELETE_INFRASTRUCTURE_WAIT"

	case DeletionStateTypeDeleteInfrastructureSucceeded:
		enumVal = "DELETE_INFRASTRUCTURE_SUCCEEDED"

	case DeletionStateTypeDeleteInfrastructureFailed:
		enumVal = "DELETE_INFRASTRUCTURE_FAILED"
	}

	return enumVal
}

// MarshalText marshals DeletionStateType into text.
func (dst DeletionStateType) MarshalText() ([]byte, error) {
	return []byte(dst.String()), nil
}

// UnmarshalText unmarshals DeletionStateType from text.
func (dst *DeletionStateType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "NOT_DELETED":
		*dst = DeletionStateTypeNotDeleted

	case "DELETE_INFRASTRUCTURE_WAIT":
		*dst = DeletionStateTypeDeleteInfrastructureWait

	case "DELETE_INFRASTRUCTURE_SUCCEEDED":
		*dst = DeletionStateTypeDeleteInfrastructureSucceeded

	case "DELETE_INFRASTRUCTURE_FAILED":
		*dst = DeletionStateTypeDeleteInfrastructureFailed

	default:
		return errors.New("invalid DeletionStateType")
	}

	return nil
}

// Value satisfies the sql/driver.Valuer interface for DeletionStateType.
func (dst DeletionStateType) Value() (driver.Value, error) {
	return dst.String(), nil
}

// Scan satisfies the database/sql.Scanner interface for DeletionStateType.
func (dst *DeletionStateType) Scan(src interface{}) error {
	buf, ok := src.([]byte)
	if !ok {
		return errors.New("invalid DeletionStateType")
	}

	return dst.UnmarshalText(buf)
}
