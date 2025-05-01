CREATE TABLE people (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	surname TEXT NOT NULL,
	patronymic TEXT,
	gender TEXT,
	age INT,
	nationality TEXT
);