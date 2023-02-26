CREATE TABLE user (
	user_id INTEGER PRIMARY KEY,
	name TEXT,
	email TEXT NOT NULL UNIQUE,
	hashed_password TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE keyboard (
	keyboard_id INTEGER PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE keycap_material (
	keycap_material_id INTEGER PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE keyswitch_type (
	keyswitch_type_id INTEGER PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE keyswitch (
	keyswitch_id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	keyswitch_type_id INTEGER NOT NULL REFERENCES keyswitch_type
);

CREATE TABLE plate_material (
	plate_material_id INTEGER PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE soundtest (
	soundtest_id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES user ON DELETE CASCADE,
	keyboard_id INTEGER NOT NULL REFERENCES keyboard,
	plate_material_id INTEGER NOT NULL REFERENCES plate_material,
	keycap_material_id INTEGER NOT NULL REFERENCES keycap_material,
	keyswitch_id INTEGER NOT NULL REFERENCES keyswitch,
	url TEXT NOT NULL,
	featured_on TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE soundtest_play (
	soundtest_id INTEGER NOT NULL REFERENCES soundtest,
	user_id INTEGER NOT NULL REFERENCES user,
	keyboard_id INTEGER NOT NULL REFERENCES keyboard,
	plate_material_id INTEGER NOT NULL REFERENCES plate_material,
	keycap_material_id INTEGER NOT NULL REFERENCES keycap_material,
	keyswitch_id INTEGER NOT NULL REFERENCES keyswitch,
	created_at TEXT NOT NULL,

	PRIMARY KEY(soundtest_id, user_id)
);

CREATE TABLE vote (
	soundtest_id INTEGER NOT NULL REFERENCES soundtest,
	user_id INTEGER NOT NULL REFERENCES user,
	vote_type INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL, 


	PRIMARY KEY(soundtest_id, user_id),
	CHECK(vote_type IN (-1, 0, 1))
);
