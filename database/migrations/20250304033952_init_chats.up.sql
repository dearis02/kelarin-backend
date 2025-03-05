DO $$
BEGIN
    CREATE TYPE chat_message_content_type AS ENUM (
        'text', 
        'image',
        'video'
    );
    EXCEPTION WHEN duplicate_object THEN 
        RAISE NOTICE 'chat_message_content_type type already exists';
END $$;

CREATE TABLE IF NOT EXISTS chat_rooms (
    id UUID PRIMARY KEY,
    service_id UUID,
    offer_id UUID,
    created_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (service_id) REFERENCES services(id),
    FOREIGN KEY (offer_id) REFERENCES offers(id)
);

CREATE TABLE IF NOT EXISTS chat_room_users (
    chat_room_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (chat_room_id, user_id)
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY,
    chat_room_id UUID NOT NULL,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    content_type chat_message_content_type NOT NULL,
    read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (chat_room_id) REFERENCES chat_rooms(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);