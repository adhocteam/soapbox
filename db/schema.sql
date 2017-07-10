create type app_type as enum ('server', 'cronjob');

create table applications (
       -- TODO(paulsmith): use app generated ID
       id serial not null primary key,
       name text not null,
       type app_type not null,
       slug text,
       description text,
       internal_dns text,
       external_dns text,
       github_repo_url text,
       dockerfile_path text,
       entrypoint_override text,
       created_at timestamp with time zone not null default now(),
       updated_at timestamp with time zone not null default now()
);
