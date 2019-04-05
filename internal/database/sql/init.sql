DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS threads;
DROP TABLE IF EXISTS forum;
DROP TABLE IF EXISTS users;




CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users ( 
	nick_name citext NOT NULL CONSTRAINT firstkey PRIMARY KEY, 
	about text, 
	email citext NOT NULL UNIQUE, 
	full_name citext NOT NULL
);

CREATE TABLE forum ( 
	posts bigint NOT NULL default 0, 
	slug text NOT NULL CONSTRAINT first_key PRIMARY KEY, 
	threads integer NOT NULL default 0, 
	title citext NOT NULL,
	user_nick citext NOT NULL, 
	CONSTRAINT forum_foreignkey FOREIGN KEY (user_nick) 
	REFERENCES users (nick_name) ON UPDATE CASCADE ON DELETE NO ACTION 
 ); 

CREATE TABLE threads ( 
	id serial CONSTRAINT threads_first_key PRIMARY KEY, 
	author citext NOT NULL, 
	CONSTRAINT thread_user_foreignkey FOREIGN KEY (author) 
	REFERENCES users (nick_name) ON UPDATE CASCADE ON DELETE NO ACTION, 
	created TIMESTAMP default now(), 
	forum text, 
	CONSTRAINT thread_forum_fk FOREIGN KEY (forum) 
	REFERENCES forum (slug) ON UPDATE CASCADE ON DELETE NO ACTION,
	message text NOT NULL,
	slug text,
	title citext NOT NULL,
	votes integer default 0
);

CREATE TABLE posts ( 
	id bigserial CONSTRAINT post_pk PRIMARY KEY, 
	author citext NOT NULL,  
	CONSTRAINT posts_user_foreignkey FOREIGN KEY (author)  
	REFERENCES users (nick_name) ON UPDATE CASCADE ON DELETE NO ACTION, 
	created TIMESTAMP default now(),
	forum text,
	CONSTRAINT posts_forum_fk FOREIGN KEY (forum)
	REFERENCES forum (slug) ON UPDATE CASCADE ON DELETE NO ACTION,
	message text NOT NULL, 
	parent bigint default 0, 
	thread integer,
	is_edited boolean default true,
	CONSTRAINT posts_thread_foreignkey FOREIGN KEY (thread)
	REFERENCES threads (id) ON UPDATE CASCADE ON DELETE NO ACTION 
); 

CREATE TABLE votes ( 
	vote_id bigserial CONSTRAINT votes_pk PRIMARY KEY, 
	thread integer,  
	CONSTRAINT votes_thread_foreignkey FOREIGN KEY (thread)
	REFERENCES threads (id) ON UPDATE CASCADE ON DELETE NO ACTION, 
	author citext NOT NULL,  
	CONSTRAINT votes_user_foreignkey FOREIGN KEY (author)  
	REFERENCES users (nick_name) ON UPDATE CASCADE ON DELETE NO ACTION, 
	is_like boolean default true, 
	CONSTRAINT user_thread_unique UNIQUE(author, thread) 
 );