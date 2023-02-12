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

CREATE TABLE sound_test (
	sound_test_id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users ON DELETE CASCADE,
	keyboard_id INTEGER NOT NULL REFERENCES keyboard,
	plate_material_id INTEGER NOT NULL REFERENCES plate_material,
	keycap_material_id INTEGER NOT NULL REFERENCES keycap_material,
	keyswitch_id INTEGER NOT NULL REFERENCES keyswitch,
	url TEXT NOT NULL,
	featured_on TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE sound_test_play (
	sound_test_id INTEGER NOT NULL REFERENCES sound_test,
	user_id INTEGER NOT NULL REFERENCES user,
	keyboard_id INTEGER NOT NULL REFERENCES keyboard,
	plate_material_id INTEGER NOT NULL REFERENCES plate_material,
	keycap_material_id INTEGER NOT NULL REFERENCES keycap_material,
	keyswitch_id INTEGER NOT NULL REFERENCES keyswitch,
	created_at TEXT NOT NULL,

	PRIMARY KEY(sound_test_id, user_id)
);

CREATE TABLE vote (
	sound_test_id INTEGER NOT NULL REFERENCES sound_test,
	user_id INTEGER NOT NULL REFERENCES user,
	vote_type INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL, 


	PRIMARY KEY(sound_test_id, user_id),
	CHECK(vote_type IN (-1, 0, 1))
);
