CREATE TABLE segments (
        id SERIAL PRIMARY KEY,
        segment_name TEXT
);

CREATE TABLE user_segments (
       user_id INT,
       segment_id INT REFERENCES segments(id),
       PRIMARY KEY (user_id, segment_id)
);