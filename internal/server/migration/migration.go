package migration

var Versions = []string{
	`CREATE TABLE IF NOT EXISTS jobs (
		id uuid,
		payload bytea,
		created_at timestamp default now(),
		activity_id uuid,
		job_type int,
		is_completed boolean default false,
		parent_job_id uuid,
		PRIMARY KEY ( id )
	);
	
	CREATE TABLE IF NOT EXISTS workers (
		id uuid,
		payload bytea,
		created_at timestamp default now(),
		PRIMARY KEY ( id )
	);
	
	CREATE TABLE IF NOT EXISTS activities (
		id uuid,
		payload bytea,
		created_at timestamp default now(),
		PRIMARY KEY ( id )
	);
	
	CREATE TABLE IF NOT EXISTS job_worker_mappings (
		id serial not null,
		job_id uuid,
		worker_id uuid,
		created_at timestamp default now(),
		PRIMARY KEY ( id )
	);
	
	CREATE TABLE IF NOT EXISTS activity_worker_mappings (
		id serial not null,
		activity_id uuid,
		worker_id uuid,
		created_at timestamp default now(),
		PRIMARY KEY ( id )
	);
	
	CREATE TABLE IF NOT EXISTS job_logs (
		id serial not null,
		job_id uuid,
		timestamp int,
		pb bytea,
		PRIMARY KEY ( id )
	);
	`,
}
