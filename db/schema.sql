create type app_type as enum ('server', 'cronjob');
create type creation_state_type as enum (
       'CREATE_INFRASTRUCTURE_WAIT',
       'SUCCEEDED',
       'FAILED');
create type deployment_state_type as enum (
       'DEPLOYMENT_ROLLOUT_WAIT',
       'DEPLOYMENT_EVALUATE_WAIT',
       'DEPLOYMENT_ROLL_FORWARD',
       'DEPLOYMENT_SUCCEEDED',
       'DEPLOYMENT_FAILED');

create table applications (
       -- TODO(paulsmith): use app generated ID
       id serial not null primary key,
       name text not null,
       type app_type not null,
       slug text not null,
       description text,
       internal_dns text,
       external_dns text,
       github_repo_url text,
       dockerfile_path text,
       entrypoint_override text,
       creation_state creation_state_type not null default 'CREATE_INFRASTRUCTURE_WAIT',
       created_at timestamp with time zone not null default now(),
       updated_at timestamp with time zone not null default now()
);

create table environments (
       id serial not null primary key,
       application_id integer references applications on delete cascade,
       name text not null,
       slug text not null,
       vars jsonb not null default '[]',
       created_at timestamp with time zone not null default now(),
       unique (application_id, name),
       unique (application_id, slug)
);

create table deployments (
       id serial not null primary key,
       application_id integer references applications on delete cascade,
       environment_id integer references environments on delete cascade,
       committish text not null,
       -- TODO(paulsmith): enum? some type safety on valid values of 'current_state'?
       current_state deployment_state_type not null default 'DEPLOYMENT_ROLLOUT_WAIT',
       created_at timestamp with time zone not null default now()
);
