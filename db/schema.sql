create type app_type as enum ('server', 'cronjob');
create type creation_state_type as enum (
  'CREATE_INFRASTRUCTURE_WAIT',
  'CREATE_INFRASTRUCTURE_SUCCEEDED',
  'CREATE_INFRASTRUCTURE_FAILED'
 );

create type deletion_state_type as enum (
  'NOT_DELETED',
  'DELETE_INFRASTRUCTURE_WAIT',
  'DELETE_INFRASTRUCTURE_SUCCEEDED',
  'DELETE_INFRASTRUCTURE_FAILED'
);

create table users (
  id serial not null primary key,
  name text not null,
  email text not null,
  encrypted_password text not null,
  github_oauth_access_token text not null default '',
  created_at timestamp with time zone not null default now(),
  updated_at timestamp with time zone not null default now(),
  unique (email)
);

create table applications (
  -- TODO(paulsmith): use app generated ID
  id serial not null primary key,
  user_id integer not null references users,
  name text not null,
  type app_type not null,
  slug text not null,
  description text,
  internal_dns text,
  external_dns text,
  github_repo_url text,
  dockerfile_path text,
  entrypoint_override text,
  aws_encryption_key_arn text not null default '',
  creation_state creation_state_type not null default 'CREATE_INFRASTRUCTURE_WAIT',
  deletion_state deletion_state_type not null default 'NOT_DELETED',
  created_at timestamp with time zone not null default now(),
  updated_at timestamp with time zone not null default now(),
  deleted_at timestamp with time zone
);

create table environments (
  id serial not null primary key,
  application_id integer references applications on delete cascade,
  name text not null,
  slug text not null,
  created_at timestamp with time zone not null default now(),
  unique (application_id, name),
  unique (application_id, slug)
);

create table configurations (
  environment_id integer references environments on delete cascade,
  version integer generated always as identity,
  created_at timestamp with time zone not null default now(),
  unique (environment_id, version)
);

create table deployments (
  id serial not null primary key,
  application_id integer references applications on delete cascade,
  environment_id integer references environments on delete cascade,
  committish text not null,
  -- TODO(paulsmith): enum? some type safety on valid values of 'current_state'?
  current_state text not null default '',
  created_at timestamp with time zone not null default now()
);

create type activity_type as enum (
  'application_created',
  'application_deleted',
  'deployment_started',
  'deployment_success',
  'deployment_failure',
  'environment_created',
  'environment_destroyed'
);

create table activities (
  id serial not null primary key,
  user_id integer references users,
  activity activity_type,
  application_id integer references applications,
  deployment_id integer references deployments,
  environment_id integer references environments,
  created_at timestamp with time zone not null default now()
);
