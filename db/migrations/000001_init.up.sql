-- Support to auto-generate UUIDs (aka GUIDs)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
--
-- Email limit is 254 chars according to RFC 5321.
-- Password is encrypted using bcrypt, which is always has 60 chars.
CREATE TABLE "users" (
    "id" uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    "email" VARCHAR(254) UNIQUE NOT NULL,
    "name" VARCHAR(64) NOT NULL,
    "password" CHAR(60) NOT NULL
);

-- User email index, email lookup is frequently used by auth
CREATE UNIQUE INDEX user_emails_idx ON users(email);

-- Groups table
--
-- Each group is owned by one user
CREATE TABLE "groups" (
    "id" uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    "name" VARCHAR(64) NOT NULL,
    "owner_id" uuid NOT NULL,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Group members table
CREATE TABLE "group_members" (
    "id" uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    "group_id" uuid NOT NULL,
    "member_id" uuid NOT NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES users(id) ON DELETE CASCADE
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
CREATE TABLE "loans" (
    "id" uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    "lender_id" uuid NOT NULL,
    "debtor_id" uuid NOT NULL,
    "amount" integer NOT NULL DEFAULT 0,

    -- TODO: select proper on delete strategy
    FOREIGN KEY (lender_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (debtor_id) REFERENCES users(id) ON DELETE CASCADE
);
