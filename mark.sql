create table users
(
    id       bigserial primary key,
    about    varchar(500),
    email    varchar(200),
    fullname varchar(200),
    nickname varchar(200)
);

create table forum
(
    id       bigserial primary key,
    posts    int,
    slug     varchar(80),
    threads  int,
    title    varchar(200),
    user_id  int references users (id) on update cascade on delete cascade
);

create table thread
(
    id          bigserial primary key,
    created     timestamp WITH TIME ZONE,
    message     varchar(3000),
    title       varchar(200),
    votes       int,

    slug       varchar(200),
    forum       varchar(200),

    forum_id int references forum (id) on update cascade on delete cascade,
    user_id  int references users (id) on update cascade on delete cascade,

    users_nickname     varchar(80),
    users_fullname     varchar(80),
    users_email        varchar(80),
    users_about        varchar(500)
);

create table post
(
    id         bigserial primary key,
    created    timestamp WITH TIME ZONE,
    forum      varchar(80),
    isEdited   boolean,
    message    varchar(5000),
    parent     int,

    thread_id  int references thread (id) on update cascade on delete cascade,
    user_id    int references users (id) on update cascade on delete cascade,

    users_nickname     varchar(80),
    users_fullname     varchar(80),
    users_email        varchar(80),
    users_about        varchar(500)
);

create table vote
(
    id       bigserial primary key,
    voice    int,
    thread_id int,
    nickname varchar(80)
);