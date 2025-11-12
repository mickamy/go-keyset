CREATE DATABASE keyset;

\c keyset

CREATE TABLE IF NOT EXISTS posts
(
    id         BIGSERIAL PRIMARY KEY,
    title      TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_posts_created_at ON posts (created_at);

INSERT INTO posts (title)
VALUES ('Post #1'),
       ('Post #2'),
       ('Post #3'),
       ('Post #4'),
       ('Post #5'),
       ('Post #6'),
       ('Post #7'),
       ('Post #8'),
       ('Post #9'),
       ('Post #10'),
       ('Post #11'),
       ('Post #12'),
       ('Post #13'),
       ('Post #14'),
       ('Post #15'),
       ('Post #16'),
       ('Post #17'),
       ('Post #18'),
       ('Post #19'),
       ('Post #20'),
       ('Post #21'),
       ('Post #22'),
       ('Post #23'),
       ('Post #24'),
       ('Post #25'),
       ('Post #26'),
       ('Post #27'),
       ('Post #28'),
       ('Post #29'),
       ('Post #30');
