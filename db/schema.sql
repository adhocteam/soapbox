create type app_type as enum ('server', 'cronjob');
create type creation_state_type as enum (
       'CREATE_INFRASTRUCTURE_WAIT',
       'SUCCEEDED',
       'FAILED');

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
       creation_state creation_state_type not null default 'CREATE_INFRASTRUCTURE_WAIT',
       created_at timestamp with time zone not null default now(),
       updated_at timestamp with time zone not null default now()
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
       version integer not null default 1, -- TODO(paulsmith): might want to do auto-incrementing here with a sequence + trigger
       created_at timestamp with time zone not null default now(),
       unique (environment_id, version)
);

create table config_vars (
       environment_id integer references environments on delete cascade,
       version integer not null,
       name text not null,
       value text not null,
       unique (environment_id, version, name)
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
        'deployment_started',
        'deployment_success',
        'deployment_failure',
        'environment_created',
        'environment_destroyed');

create table activities (
        id serial not null primary key,
        user_id integer references users,
        activity activity_type,
        application_id integer references applications,
        deployment_id integer references deployments,
        environment_id integer references environments,
        created_at timestamp with time zone not null default now()
);

-- deploy state machine

create table deploy_events (
       id serial not null primary key,
       deploy_id integer references deployments on delete cascade,
       event text not null,
       created_at timestamp with time zone not null default now()
);

create function deploy_events_transition(state text, event text) returns text
language sql as
$$
select case state
   when 'start' then
       case event
       	    when 'rollout-started' then 'rollout-wait'
	    else 'error'
       end
   when 'rollout-wait' then
       case event
       	    when 'rollout-in-progress' then 'rollout-wait'
	    when 'rollout-ok' then 'evaluate-wait'
	    when 'rollout-failed' then 'rollback'
	    else 'error'
       end
   when 'evaluate-wait' then
       case event
       	    when 'evaluate-in-progress' then 'evaluate-wait'
       	    when 'evaluate-ok' then 'rollforward'
       	    when 'evaluate-failed' then 'rollback'
	    else 'error'
       end
   when 'rollforward' then
       case event
       	    when 'rollforward-started' then 'rollforward-wait'
	    else 'error'
       end
   when 'rollforward-wait' then
       case event
       	    when 'rollforward-in-progress' then 'rollforward-wait'
       	    when 'rollforward-ok' then 'success'
	    else 'error'
       end
   when 'rollback' then
       case event
       	    when 'rollback-started' then 'rollback-wait'
	    else 'error'
       end
   when 'rollback-wait' then
       case event
       	    when 'rollback-in-progress' then 'rollback-wait'
       	    when 'rollback-ok' then 'failure'
	    else 'error'
       end
   else 'error'
end  
$$;

create aggregate deploy_events_fsm(text) (
    sfunc = deploy_events_transition,
    stype = text,
    initcond = 'start'
);

create function deploy_events_trigger_func() returns trigger
language plpgsql as $$
declare
    new_state text;
begin
    select deploy_events_fsm(event order by id)
    from (
        select id, event from deploy_events where deploy_id = new.deploy_id
	union
	select new.id, new.event
    ) s
    into new_state;

    if new_state = 'error' then
        raise exception 'invalid event';
    end if;

    return new;
end
$$;

create trigger deploy_events_trigger before insert on deploy_events
for each row execute procedure deploy_events_trigger_func();
