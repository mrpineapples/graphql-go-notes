create extension pgcrypto;

-- NOTE:
--
-- 1:
--
-- 'u-' || substr(gen_random_uuid()::text, 1, 6) generates
-- a random user ID with a pattern like the following:
--
-- u-a1bbe1
-- u-92f31c
-- u-2501e9
--
-- 2:
--
-- check (username ~ '^\w{3,8}$') checks that a username is
-- 3-8 alphanumeric characters.

create table users (
  user_id  text not null unique default 'u-' || substr(gen_random_uuid()::text, 1, 6),
  username text not null unique check (username ~ '^\w{3,15}$') );

create table notes (
  user_id  text not null references users (user_id),
  note_id  text not null unique default 'n-' || substr(gen_random_uuid()::text, 1, 6),
  data     text not null );

-- Insert mock data:
insert into users (username) values ('CR7');
insert into users (username) values ('SergioRamos');
insert into users (username) values ('michaelmiranda');

insert into notes (user_id, data) values ((select user_id from users where username = 'CR7'), 'Olá Mundo!');
insert into notes (user_id, data) values ((select user_id from users where username = 'CR7'), 'Olá novamente, mundo!');
insert into notes (user_id, data) values ((select user_id from users where username = 'CR7'), 'Olá, escuridão!');
insert into notes (user_id, data) values ((select user_id from users where username = 'SergioRamos' ), '!Hola Mundo!');
insert into notes (user_id, data) values ((select user_id from users where username = 'SergioRamos' ), '¡Hola de nuevo mundo!');
insert into notes (user_id, data) values ((select user_id from users where username = 'SergioRamos' ), '¡Hola oscuridad!');
insert into notes (user_id, data) values ((select user_id from users where username = 'michaelmiranda' ), 'Hello, world!');
insert into notes (user_id, data) values ((select user_id from users where username = 'michaelmiranda' ), 'Hello again, world!');
insert into notes (user_id, data) values ((select user_id from users where username = 'michaelmiranda' ), 'Hello, darkness!');
