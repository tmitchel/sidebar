DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE users (
    id SERIAL UNIQUE,
    display_name VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    profile_image TEXT NOT NULL,
    PRIMARY KEY(id)
);

DROP TABLE IF EXISTS messages CASCADE;
CREATE TABLE messages (
    id SERIAL UNIQUE,
    content TEXT NOT NULL,
    event INT NOT NULL,
    PRIMARY KEY(id)
);

DROP TABLE IF EXISTS channels CASCADE;
CREATE TABLE channels (
    id SERIAL UNIQUE,
    display_name VARCHAR(255) UNIQUE NOT NULL,
    is_sidebar BOOLEAN DEFAULT FALSE,
    is_direct BOOLEAN DEFAULT FALSE,
    resolved BOOLEAN DEFAULT FALSE,
    PRIMARY KEY(id)
);

DROP TABLE IF EXISTS sidebars;
CREATE TABLE sidebars (
    id INT NOT NULL,
    parent_id INT,
    FOREIGN KEY(id) REFERENCES channels(id) ON DELETE CASCADE,
    FOREIGN KEY(parent_id) REFERENCES channels(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS users_channels CASCADE;
CREATE TABLE users_channels (
    user_id INT REFERENCES users (id) ON UPDATE CASCADE,
    channel_id INT REFERENCES channels (id) ON UPDATE CASCADE,
    CONSTRAINT users_channels_pkey PRIMARY KEY (user_id, channel_id)
);

DROP TABLE IF EXISTS tokens CASCADE;
CREATE TABLE tokens (
    token VARCHAR(255) NOT NULL UNIQUE,
    creater_id INT NOT NULL,
    new_user_id INT,
    valid BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(token),
    FOREIGN KEY(creater_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(new_user_id) REFERENCES users(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS users_messages CASCADE;
CREATE TABLE users_messages (
    user_to_id INT,
    user_from_id INT NOT NULL,
    message_id INT NOT NULL,
    FOREIGN KEY(user_to_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(user_from_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS channels_messages CASCADE;
CREATE TABLE channels_messages (
    channel_id INT NOT NULL,
    message_id INT NOT NULL,
    FOREIGN KEY(channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE
);