CREATE TABLE IF NOT EXISTS jobs (
    id uuid,
    payload bytea,
	PRIMARY KEY ( id )
);

CREATE TABLE IF NOT EXISTS workers (
    id uuid,
    payload bytea,
	PRIMARY KEY ( id )
);

CREATE TABLE IF NOT EXISTS job_worker_mappings (
    id serial not null,
    job_id uuid,
    activity_id uuid,
    worker_id uuid,
	PRIMARY KEY ( id )
);