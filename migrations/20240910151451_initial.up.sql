CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE service_type AS ENUM (
    'Construction',
    'Delivery',
    'Manufacture'
    );

CREATE TYPE tender_status AS ENUM (
    'Created',
    'Published',
    'Closed'
    );

CREATE TYPE bid_status AS ENUM (
    'Created',
    'Published',
    'Canceled'
    );

CREATE TYPE author_type AS ENUM (
    'User',
    'Organization'
    );

CREATE TABLE tender
(
    id               UUID          NOT NULL DEFAULT uuid_generate_v4(),
    name             VARCHAR(100)  NOT NULL,
    description      TEXT          NOT NULL,
    type             service_type  NOT NULL,
    status           tender_status NOT NULL DEFAULT 'Created',
    organization_id  UUID          NOT NULL REFERENCES organization (id) ON DELETE CASCADE,
    version          INT           NOT NULL DEFAULT 1,
    creator_username VARCHAR(50)   NOT NULL,
    created_at       TIMESTAMP              DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP              DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, version)
);

CREATE TABlE bid
(
    id          UUID         NOT NULL DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL,
    description TEXT         NOT NULL,
    tender_id   UUID         NOT NULL,
    status      bid_status   NOT NULL DEFAULT 'Created',
    decision    VARCHAR(20),
    author_type author_type  NOT NULL,
    author_id   UUID         NOT NULL,
    version     INT          NOT NULL DEFAULT 1,
    created_at  TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, version)
);

CREATE INDEX idx_tender_id_hash ON tender USING HASH (id);
CREATE INDEX idx_bid_id_hash ON bid Using HASH (id);
CREATE INDEX idx_tender_creator_username_hash ON tender USING HASH (creator_username);