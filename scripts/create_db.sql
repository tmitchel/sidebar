DROP TABLE IF EXISTS workspaces CASCADE;
CREATE TABLE workspaces (
    id VARCHAR(36) UNIQUE NOT NULL,
    token VARCHAR(36) UNIQUE NOT NULL,
    display_name VARCHAR(255) UNIQUE NOT NULL,
    display_image TEXT NOT NULL,
    default_ws BOOLEAN DEFAULT FALSE,
    PRIMARY KEY(id)
);

DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE users (
    id VARCHAR(36) UNIQUE NOT NULL,
    display_name VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    profile_image TEXT NOT NULL,
    user_role INT NOT NULL DEFAULT 1,
    PRIMARY KEY(id)
);

DROP TABLE IF EXISTS channels CASCADE;
CREATE TABLE channels (
    id VARCHAR(36) UNIQUE,
    display_name VARCHAR(255) UNIQUE NOT NULL,
    details TEXT,
    display_image TEXT NOT NULL,
    is_sidebar BOOLEAN DEFAULT FALSE,
    is_direct BOOLEAN DEFAULT FALSE,
    resolved BOOLEAN DEFAULT FALSE,
    PRIMARY KEY(id)
);

DROP TABLE IF EXISTS workspaces_users CASCADE;
CREATE TABLE workspaces_users (
    workspace_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    FOREIGN KEY(workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS workspaces_channels CASCADE;
CREATE TABLE workspaces_channels (
    workspace_id VARCHAR(36) NOT NULL,
    channel_id VARCHAR(36) NOT NULL,
    FOREIGN KEY(workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    FOREIGN KEY(channel_id) REFERENCES channels(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS messages CASCADE;
CREATE TABLE messages (
    id VARCHAR(36) UNIQUE NOT NULL,
    content TEXT NOT NULL,
    event INT NOT NULL,
    PRIMARY KEY(id)
);

DROP TABLE IF EXISTS sidebars;
CREATE TABLE sidebars (
    id VARCHAR(36) NOT NULL,
    parent_id VARCHAR(36),
    FOREIGN KEY(id) REFERENCES channels(id) ON DELETE CASCADE,
    FOREIGN KEY(parent_id) REFERENCES channels(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS users_channels CASCADE;
CREATE TABLE users_channels (
    user_id VARCHAR(36) REFERENCES users (id) ON UPDATE CASCADE,
    channel_id VARCHAR(36) REFERENCES channels (id) ON UPDATE CASCADE,
    CONSTRAINT users_channels_pkey PRIMARY KEY (user_id, channel_id)
);

DROP TABLE IF EXISTS tokens CASCADE;
CREATE TABLE tokens (
    token VARCHAR(255) NOT NULL UNIQUE,
    creater_id VARCHAR(36) NOT NULL,
    new_user_id VARCHAR(36),
    valid BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(token),
    FOREIGN KEY(creater_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(new_user_id) REFERENCES users(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS users_messages CASCADE;
CREATE TABLE users_messages (
    user_to_id VARCHAR(36),
    user_from_id VARCHAR(36) NOT NULL,
    message_id VARCHAR(36) NOT NULL,
    FOREIGN KEY(user_from_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS channels_messages CASCADE;
CREATE TABLE channels_messages (
    channel_id VARCHAR(36) NOT NULL,
    message_id VARCHAR(36) NOT NULL,
    FOREIGN KEY(channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE
);