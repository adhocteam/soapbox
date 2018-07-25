// Package models contains the types for schema 'public'.
package models

// GENERATED BY XO. DO NOT EDIT.

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

// Application represents a row from 'public.applications'.
type Application struct {
	ID                  int               `json:"id"`                     // id
	UserID              int               `json:"user_id"`                // user_id
	Name                string            `json:"name"`                   // name
	Type                AppType           `json:"type"`                   // type
	Slug                string            `json:"slug"`                   // slug
	Description         sql.NullString    `json:"description"`            // description
	InternalDNS         sql.NullString    `json:"internal_dns"`           // internal_dns
	ExternalDNS         sql.NullString    `json:"external_dns"`           // external_dns
	GithubRepoURL       sql.NullString    `json:"github_repo_url"`        // github_repo_url
	DockerfilePath      sql.NullString    `json:"dockerfile_path"`        // dockerfile_path
	EntrypointOverride  sql.NullString    `json:"entrypoint_override"`    // entrypoint_override
	AwsEncryptionKeyArn string            `json:"aws_encryption_key_arn"` // aws_encryption_key_arn
	CreationState       CreationStateType `json:"creation_state"`         // creation_state
	DeletionState       DeletionStateType `json:"deletion_state"`         // deletion_state
	CreatedAt           time.Time         `json:"created_at"`             // created_at
	UpdatedAt           time.Time         `json:"updated_at"`             // updated_at
	DeletedAt           pq.NullTime       `json:"deleted_at"`             // deleted_at

	// xo fields
	_exists, _deleted bool
}

// Exists determines if the Application exists in the database.
func (a *Application) Exists() bool {
	return a._exists
}

// Deleted provides information if the Application has been deleted from the database.
func (a *Application) Deleted() bool {
	return a._deleted
}

// Insert inserts the Application to the database.
func (a *Application) Insert(db XODB) error {
	var err error

	// if already exist, bail
	if a._exists {
		return errors.New("insert failed: already exists")
	}

	// sql insert query, primary key provided by sequence
	const sqlstr = `INSERT INTO public.applications (` +
		`user_id, name, type, slug, description, internal_dns, external_dns, github_repo_url, dockerfile_path, entrypoint_override, aws_encryption_key_arn, creation_state, deletion_state, created_at, updated_at, deleted_at` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16` +
		`) RETURNING id`

	// run query
	XOLog(sqlstr, a.UserID, a.Name, a.Type, a.Slug, a.Description, a.InternalDNS, a.ExternalDNS, a.GithubRepoURL, a.DockerfilePath, a.EntrypointOverride, a.AwsEncryptionKeyArn, a.CreationState, a.DeletionState, a.CreatedAt, a.UpdatedAt, a.DeletedAt)
	err = db.QueryRow(sqlstr, a.UserID, a.Name, a.Type, a.Slug, a.Description, a.InternalDNS, a.ExternalDNS, a.GithubRepoURL, a.DockerfilePath, a.EntrypointOverride, a.AwsEncryptionKeyArn, a.CreationState, a.DeletionState, a.CreatedAt, a.UpdatedAt, a.DeletedAt).Scan(&a.ID)
	if err != nil {
		return err
	}

	// set existence
	a._exists = true

	return nil
}

// Update updates the Application in the database.
func (a *Application) Update(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !a._exists {
		return errors.New("update failed: does not exist")
	}

	// if deleted, bail
	if a._deleted {
		return errors.New("update failed: marked for deletion")
	}

	// sql query
	const sqlstr = `UPDATE public.applications SET (` +
		`user_id, name, type, slug, description, internal_dns, external_dns, github_repo_url, dockerfile_path, entrypoint_override, aws_encryption_key_arn, creation_state, deletion_state, created_at, updated_at, deleted_at` +
		`) = ( ` +
		`$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16` +
		`) WHERE id = $17`

	// run query
	XOLog(sqlstr, a.UserID, a.Name, a.Type, a.Slug, a.Description, a.InternalDNS, a.ExternalDNS, a.GithubRepoURL, a.DockerfilePath, a.EntrypointOverride, a.AwsEncryptionKeyArn, a.CreationState, a.DeletionState, a.CreatedAt, a.UpdatedAt, a.DeletedAt, a.ID)
	_, err = db.Exec(sqlstr, a.UserID, a.Name, a.Type, a.Slug, a.Description, a.InternalDNS, a.ExternalDNS, a.GithubRepoURL, a.DockerfilePath, a.EntrypointOverride, a.AwsEncryptionKeyArn, a.CreationState, a.DeletionState, a.CreatedAt, a.UpdatedAt, a.DeletedAt, a.ID)
	return err
}

// Save saves the Application to the database.
func (a *Application) Save(db XODB) error {
	if a.Exists() {
		return a.Update(db)
	}

	return a.Insert(db)
}

// Upsert performs an upsert for Application.
//
// NOTE: PostgreSQL 9.5+ only
func (a *Application) Upsert(db XODB) error {
	var err error

	// if already exist, bail
	if a._exists {
		return errors.New("insert failed: already exists")
	}

	// sql query
	const sqlstr = `INSERT INTO public.applications (` +
		`id, user_id, name, type, slug, description, internal_dns, external_dns, github_repo_url, dockerfile_path, entrypoint_override, aws_encryption_key_arn, creation_state, deletion_state, created_at, updated_at, deleted_at` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17` +
		`) ON CONFLICT (id) DO UPDATE SET (` +
		`id, user_id, name, type, slug, description, internal_dns, external_dns, github_repo_url, dockerfile_path, entrypoint_override, aws_encryption_key_arn, creation_state, deletion_state, created_at, updated_at, deleted_at` +
		`) = (` +
		`EXCLUDED.id, EXCLUDED.user_id, EXCLUDED.name, EXCLUDED.type, EXCLUDED.slug, EXCLUDED.description, EXCLUDED.internal_dns, EXCLUDED.external_dns, EXCLUDED.github_repo_url, EXCLUDED.dockerfile_path, EXCLUDED.entrypoint_override, EXCLUDED.aws_encryption_key_arn, EXCLUDED.creation_state, EXCLUDED.deletion_state, EXCLUDED.created_at, EXCLUDED.updated_at, EXCLUDED.deleted_at` +
		`)`

	// run query
	XOLog(sqlstr, a.ID, a.UserID, a.Name, a.Type, a.Slug, a.Description, a.InternalDNS, a.ExternalDNS, a.GithubRepoURL, a.DockerfilePath, a.EntrypointOverride, a.AwsEncryptionKeyArn, a.CreationState, a.DeletionState, a.CreatedAt, a.UpdatedAt, a.DeletedAt)
	_, err = db.Exec(sqlstr, a.ID, a.UserID, a.Name, a.Type, a.Slug, a.Description, a.InternalDNS, a.ExternalDNS, a.GithubRepoURL, a.DockerfilePath, a.EntrypointOverride, a.AwsEncryptionKeyArn, a.CreationState, a.DeletionState, a.CreatedAt, a.UpdatedAt, a.DeletedAt)
	if err != nil {
		return err
	}

	// set existence
	a._exists = true

	return nil
}

// Delete deletes the Application from the database.
func (a *Application) Delete(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !a._exists {
		return nil
	}

	// if deleted, bail
	if a._deleted {
		return nil
	}

	// sql query
	const sqlstr = `DELETE FROM public.applications WHERE id = $1`

	// run query
	XOLog(sqlstr, a.ID)
	_, err = db.Exec(sqlstr, a.ID)
	if err != nil {
		return err
	}

	// set deleted
	a._deleted = true

	return nil
}

// User returns the User associated with the Application's UserID (user_id).
//
// Generated from foreign key 'applications_user_id_fkey'.
func (a *Application) User(db XODB) (*User, error) {
	return UserByID(db, a.UserID)
}

// ApplicationByID retrieves a row from 'public.applications' as a Application.
//
// Generated from index 'applications_pkey'.
func ApplicationByID(db XODB, id int) (*Application, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, user_id, name, type, slug, description, internal_dns, external_dns, github_repo_url, dockerfile_path, entrypoint_override, aws_encryption_key_arn, creation_state, deletion_state, created_at, updated_at, deleted_at ` +
		`FROM public.applications ` +
		`WHERE id = $1`

	// run query
	XOLog(sqlstr, id)
	a := Application{
		_exists: true,
	}

	err = db.QueryRow(sqlstr, id).Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Slug, &a.Description, &a.InternalDNS, &a.ExternalDNS, &a.GithubRepoURL, &a.DockerfilePath, &a.EntrypointOverride, &a.AwsEncryptionKeyArn, &a.CreationState, &a.DeletionState, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt)
	if err != nil {
		return nil, err
	}

	return &a, nil
}
