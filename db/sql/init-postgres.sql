CREATE TABLE people (
    id serial PRIMARY KEY,
    name text NOT NULL,
    created_at date NOT NULL DEFAULT CURRENT_DATE
);

CREATE TABLE friendships (
    person1_id int REFERENCES people (id) ON DELETE CASCADE,
    person2_id int REFERENCES people (id) ON DELETE CASCADE
);