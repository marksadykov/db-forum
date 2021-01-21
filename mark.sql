ALTER SYSTEM SET
    checkpoint_completion_target = '0.9';
ALTER SYSTEM SET
    wal_buffers = '6912kB';
ALTER SYSTEM SET
    default_statistics_target = '100';
ALTER SYSTEM SET
    random_page_cost = '1.1';
ALTER SYSTEM SET
    effective_io_concurrency = '200';

CREATE UNLOGGED TABLE users
(
    id       SERIAL PRIMARY KEY,
    about    varchar(500),
    email    varchar(200),
    fullname varchar(200),
    nickname varchar(200)
);

CREATE UNLOGGED TABLE forum
(
    id       SERIAL PRIMARY KEY,
    posts    int,
    slug     varchar(80),
    threads  int,
    title    varchar(200),
    nickname varchar(200),
    user_id int
);

CREATE UNLOGGED TABLE thread
(
    id          SERIAL PRIMARY KEY,
    created     timestamp WITH TIME ZONE,
    message     varchar(3000),
    title       varchar(200),
    votes       int,

    slug       varchar(200),
    forum       varchar(200),

    forum_id int,
    user_id  int,

    users_nickname     varchar(80),
    users_fullname     varchar(80),
    users_email        varchar(80),
    users_about        varchar(500)
);

CREATE UNLOGGED TABLE post
(
    id         SERIAL PRIMARY KEY,
    created    timestamp WITH TIME ZONE,
    forum      varchar(80),
    isEdited   boolean,
    message    varchar(5000),
    parent     int,

    thread_id  int,
    user_id    int,

    users_nickname     varchar(80),
    users_fullname     varchar(80),
    users_email        varchar(80),
    users_about        varchar(500)
);

CREATE UNLOGGED TABLE vote
(
    id       SERIAL PRIMARY KEY,
    voice    int,
    thread_id int,
    nickname varchar(80),
    thread_slug varchar(80)
);

-- for input
CREATE INDEX users_nickname_lower_index ON users (lower(nickname));
CREATE INDEX users_email_index ON users (lower(email));
CREATE INDEX forum_slug_lower_index ON forum (lower(slug));
CREATE INDEX thread_slug_lower_index ON thread (lower(slug));
CREATE INDEX thread_id_index ON thread (id);
CREATE INDEX vote_nickname ON vote (thread_id, lower(nickname));
CREATE INDEX post_thread_id_index ON post (thread_id);

-- for output
CREATE INDEX users_id_index ON users (id);
CREATE INDEX thread_forum_id_index ON thread (forum_id);
CREATE INDEX post_id_index ON post (id);
CREATE INDEX forum_id_index ON forum (id);
CREATE INDEX post_forum_index ON post (lower(forum));