CREATE TABLE IF NOT EXISTS url (
	`key` CHAR(10) PRIMARY KEY,
	long_url TEXT NOT NULL,
	short_url TEXT NOT NULL
);
