-- Support to auto-generate UUIDs (aka GUIDs)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Here be dragons 🔥🐲 --

-- Users table
--
-- Email limit is 254 chars according to RFC 5321, don't ask me who need such long mail.
-- Password is encrypted using bcrypt, which is always has 60 chars.
CREATE TABLE "users"
(
    "id"       uuid PRIMARY KEY    NOT NULL DEFAULT uuid_generate_v4(),
    "email"    VARCHAR(254) UNIQUE NOT NULL,
    "name"     VARCHAR(64)         NOT NULL,
    "password" CHAR(60)            NOT NULL
);

-- Groups table
--
-- Each group is owned by one user
CREATE TABLE "groups"
(
    "id"       uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    "name"     VARCHAR(64)      NOT NULL,
    "owner_id" uuid             NOT NULL,
    FOREIGN KEY (owner_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Group members table
CREATE TABLE "group_membership"
(
    "group_id"  uuid NOT NULL,
    "member_id" uuid NOT NULL,
    FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES users (id) ON DELETE CASCADE,

    -- Group member add/remove will utilize member_id/group_id search
    -- so row id is mostly unnecessary.
    -- Also user can join to a group only once.
    PRIMARY KEY (group_id, member_id)
);

-- Loans table
--
-- Contains list of records about provided loan to user.
-- Lender ID is id of loan provider.
-- Debtor ID is id of debtor.
-- Amount is dept amount in cents.
--
-- Loans don't depend on group, because the same user may have
-- load to other user in multiple groups at the same time.
--
-- When user pays the bill, bill is separated between each member
-- (except who payed the bill) in group.
--
-- Load fraction ( total_bill / len(members) ) will be
-- registered as dept for each group member.
--
-- So Alice payed 4$, Bob owes her 2$:
--  (lender_id = alice, debtor_id = bob, amount=200)
CREATE TABLE "loans"
(
    "id"         uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    "lender_id"  uuid             NOT NULL,
    "debtor_id"  uuid             NOT NULL,
    "amount"     integer          NOT NULL CHECK (amount >= 0),
    "created_at" timestamptz      NOT NULL DEFAULT NOW(),

    -- I added "ON DELETE CASCADE" constraint just for DB consistency.
    -- API never removes created users from `users` table.
    FOREIGN KEY (lender_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (debtor_id) REFERENCES users (id) ON DELETE CASCADE
);

-- I intentionally didn't create any indexes for "loans".
--
-- Although we have a frequent search by "lender_id" or "debtor_id",
-- but assuming that table insert/update is a frequent operation - keeping
-- 2 indexes for huge table will bring more disadvantages than advantages
-- (update time and storage space will dramatically increase).
--
-- I could try to use composed cluster index, but:
--  1. Table reordering during update takes time
--  2. That probably won't speed search by individual field (`lender_id` or `debtor_id`)
--
-- Another trick is to write trigger or stored procedure to store
-- updated balance, but it will be to simple, isn't it?
--
-- Assuming that I do a long search query only once (on balance cache population),
-- insert time looks more important so I decided to omit index creation.
